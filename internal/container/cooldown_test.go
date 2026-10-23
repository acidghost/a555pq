package container

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/v1/remote"
)

const testDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

type hubTestTransport struct {
	t    *testing.T
	base *url.URL
}

func (tr hubTestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	allowedHost := req.URL.Host == "hub.docker.com" || req.URL.Host == "index.docker.io" || req.URL.Host == RegistryDockerV2
	if req.URL.Scheme != "https" || !allowedHost {
		err := fmt.Errorf("forbidden outbound request: %s", req.URL)
		tr.t.Error(err)
		return nil, err
	}
	original := req
	req = req.Clone(req.Context())
	req.Host = req.URL.Host
	req.URL.Scheme = tr.base.Scheme
	req.URL.Host = tr.base.Host
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	resp.Request = original
	return resp, nil
}

func newHubTestRegistry(t *testing.T, handler http.HandlerFunc) *UnifiedRegistry {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	base, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	r := NewUnifiedRegistry()
	transport := hubTestTransport{t: t, base: base}
	r.httpClient.Transport = transport
	r.options = []remote.Option{remote.WithTransport(transport)}
	return r
}

func TestDockerHubCooldown(t *testing.T) {
	old := time.Now().Add(-30 * 24 * time.Hour).UTC().Format(time.RFC3339)
	recent := time.Now().UTC().Format(time.RFC3339)
	var paths []string
	r := newHubTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		paths = append(paths, req.URL.Path)
		switch req.URL.Path {
		case "/v2/":
			fmt.Fprint(w, `{}`)
		case "/v2/library/nginx/tags/list":
			fmt.Fprint(w, `{"name":"library/nginx","tags":["1.0.0","latest","2.0.0","3.0.0","7.0.0","8.0.0","9.0.0"]}`)
		case "/v2/repositories/library/nginx/tags/":
			if req.URL.Query().Get("page") == "2" {
				fmt.Fprintf(w, `{"results":[{"name":"1.0.0","tag_last_pushed":%q,"digest":%q,"full_size":2048}]}`, old, testDigest)
				return
			}
			fmt.Fprintf(w, `{"next":"?page=2","results":[
				{"name":"3.0.0","tag_last_pushed":%q,"last_updated":%q},
				{"name":"2.0.0","tag_last_pushed":%q,"last_updated":%q,"digest":%q,"full_size":1024},
				{"name":"9.0.0","last_updated":%q},
				{"name":"8.0.0","tag_last_pushed":"invalid"},
				{"name":"7.0.0","tag_last_pushed":null},
				{"name":"latest","tag_last_pushed":%q,"digest":%q}
			]}`, recent, old, old, recent, testDigest, old, old, testDigest)
		case "/v2/repositories/library/nginx/":
			fmt.Fprint(w, `{"description":"test image"}`)
		default:
			prefix := "/v2/repositories/library/nginx/tags/"
			tag := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, prefix), "/")
			pushed := old
			switch tag {
			case "latest", "1.0.0", "2.0.0":
			case "3.0.0":
				pushed = recent
			case "7.0.0", "9.0.0":
				pushed = ""
			case "8.0.0":
				pushed = "invalid"
			case "missing":
				w.WriteHeader(http.StatusNotFound)
				return
			default:
				t.Errorf("unexpected request: %s", req.URL)
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(dockerHubTag{Name: tag, TagLastPushed: pushed, Digest: testDigest, FullSize: 1024})
		}
	})
	opts := Options{MinReleaseAge: 7 * 24 * time.Hour}
	for _, host := range []string{"", RegistryDocker, RegistryDockerV2, "index.docker.io"} {
		t.Run("registry="+host, func(t *testing.T) {
			paths = nil
			ref := ImageReference{Registry: host, Name: "nginx"}
			tags, err := r.GetTags(ref, opts)
			if err != nil {
				t.Fatal(err)
			}
			var names []string
			for _, tag := range tags.Tags {
				names = append(names, tag.Name)
			}
			if !reflect.DeepEqual(names, []string{"latest", "2.0.0", "1.0.0"}) {
				t.Fatalf("eligible tags = %v", names)
			}
			if len(paths) != 2 {
				t.Fatalf("expected one paginated scan, got %v", paths)
			}
			latest, err := r.GetLatestTag(ref, opts)
			if err != nil || latest.Name != "2.0.0" || latest.Digest != testDigest {
				t.Fatalf("latest = %+v, err = %v", latest, err)
			}
			info, err := r.GetImageInfo(ref, opts)
			if err != nil {
				t.Fatal(err)
			}
			if info.LatestTag != "2.0.0" || info.TagDate != old || info.Size != "1.0 KiB" ||
				info.Manifest == nil || info.Manifest.Digest != testDigest || !strings.HasSuffix(info.FullImageRef, "@"+testDigest) {
				t.Fatalf("show = %+v", info)
			}
			ref.Tag = "latest"
			info, err = r.GetImageInfo(ref, opts)
			if err != nil || info.LatestTag != "latest" {
				t.Fatalf("explicit old alias: info=%+v err=%v", info, err)
			}
			for _, tag := range []string{"3.0.0", "9.0.0", "8.0.0", "7.0.0", "missing"} {
				ref.Tag = tag
				if _, err := r.GetImageInfo(ref, opts); err == nil {
					t.Errorf("show %s should fail", tag)
				}
			}
		})
	}
}

func TestContainerCooldownCutoff(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	cutoff := now.Add(-24 * time.Hour)
	tags := []TagInfo{
		{Name: "old", CreatedAt: cutoff.Add(-time.Second).Format(time.RFC3339)},
		{Name: "boundary", CreatedAt: cutoff.Format(time.RFC3339)},
		{Name: "recent", CreatedAt: cutoff.Add(time.Second).Format(time.RFC3339)},
		{Name: "unknown"},
		{Name: "invalid", CreatedAt: "not a date"},
		{Name: "zero", CreatedAt: time.Time{}.Format(time.RFC3339)},
	}
	got := filterTagsByMinReleaseAge(tags, 24*time.Hour, now)
	if !reflect.DeepEqual(got, tags[:2]) {
		t.Fatalf("filtered = %+v", got)
	}
}

func TestContainerCooldownRejectsOtherRegistries(t *testing.T) {
	r := newHubTestRegistry(t, func(_ http.ResponseWriter, req *http.Request) {
		t.Errorf("unexpected request: %s", req.URL)
	})
	for _, host := range []string{GHCR, GCR, "public.ecr.aws", "example.docker.io", "docker.io.evil.example", "localhost:5000"} {
		ref := ImageReference{Registry: host, Organization: "org", Name: "image"}
		opts := Options{MinReleaseAge: time.Hour}
		if _, err := r.GetTags(ref, opts); err == nil || !strings.Contains(err.Error(), "Docker Hub") {
			t.Errorf("%s: expected unsupported registry error, got %v", host, err)
		}
		if _, err := r.GetTags(ref, Options{MaxPages: 1}); err == nil || !strings.Contains(err.Error(), "Docker Hub") {
			t.Errorf("%s: page-limited listing must reject unsupported registries, got %v", host, err)
		}
		if _, err := r.GetLatestTag(ref, opts); err == nil {
			t.Errorf("%s: latest should fail", host)
		}
		if _, err := r.GetImageInfo(ref, opts); err == nil {
			t.Errorf("%s: show should fail", host)
		}
	}
}

func TestDockerHubCooldownVersionsFailures(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"unauthorized", 401, `{}`},
		{"rate limited", 429, `{}`},
		{"server error", 500, `{}`},
		{"invalid JSON", 200, `not JSON`},
		{"empty", 200, `{"results":[]}`},
		{"unknown timestamp", 200, `{"results":[{"name":"1.0.0","last_updated":"2000-01-01T00:00:00Z"}]}`},
		{"cross origin", 200, `{"next":"https://attacker.example/tags","results":[]}`},
		{"downgrade", 200, `{"next":"http://hub.docker.com/tags","results":[]}`},
		{"loop", 200, `{"next":"?page_size=100&ordering=last_updated","results":[]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := newHubTestRegistry(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			})
			ref := ImageReference{Name: "nginx"}
			opts := Options{MinReleaseAge: time.Hour, MaxPages: 1}
			tags, err := r.GetTags(ref, opts)
			if tc.name == "empty" || tc.name == "unknown timestamp" {
				if err != nil || len(tags.Tags) != 0 {
					t.Fatalf("expected no eligible versions: tags=%+v err=%v", tags, err)
				}
			} else if err == nil || tags != nil {
				t.Fatalf("versions must not bypass failures or return partial results: tags=%+v err=%v", tags, err)
			}
		})
	}
}

func TestContainerNoCooldownUsesOCI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/v2/":
			fmt.Fprint(w, `{}`)
		case "/v2/test/tags/list":
			fmt.Fprint(w, `{"name":"test","tags":["1.0.0","2.0.0","latest"]}`)
		default:
			t.Errorf("unexpected request: %s", req.URL)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	c := NewClient()
	tag, err := c.GetLatestTag(strings.TrimPrefix(server.URL, "http://")+"/test", Options{})
	if err != nil || tag.Name != "2.0.0" || tag.Digest != "" {
		t.Fatalf("latest without cooldown = %+v, err = %v", tag, err)
	}
}

func TestDockerHubCooldownExplicitTagAvoidsPagination(t *testing.T) {
	old := time.Now().Add(-30 * 24 * time.Hour).UTC().Format(time.RFC3339)
	r := newHubTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/v2/repositories/renovate/renovate/tags/1.0.0/":
			fmt.Fprintf(w, `{"name":"1.0.0","tag_last_pushed":%q,"digest":%q}`, old, testDigest)
		case "/v2/repositories/renovate/renovate/":
			fmt.Fprint(w, `{"description":"Renovate"}`)
		default:
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"message":"must not list tags for an explicit tag"}`)
		}
	})
	info, err := (&Client{registry: r}).GetImageInfo("docker.io/renovate/renovate:1.0.0", Options{MinReleaseAge: time.Hour})
	if err != nil || info.LatestTag != "1.0.0" {
		t.Fatalf("explicit show = %+v, err = %v", info, err)
	}
}

func TestDockerHubCooldownPaginationErrorDetails(t *testing.T) {
	const message = "pagination offset too large for anonymous requests; sign in to page further"
	for _, maxPages := range []int{0, 1, 2} {
		t.Run(fmt.Sprintf("max_pages=%d", maxPages), func(t *testing.T) {
			var requests atomic.Int32
			r := newHubTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
				requests.Add(1)
				if req.URL.Query().Get("page") == "2" {
					w.WriteHeader(http.StatusForbidden)
					fmt.Fprintf(w, `{"message":%q,"errinfo":{}}`, message)
					return
				}
				fmt.Fprintf(w, `{"next":"?page=2","results":[{"name":"1.0.0","tag_last_pushed":"2000-01-01T00:00:00Z","digest":%q}]}`, testDigest)
			})
			tags, err := r.GetTags(ImageReference{Organization: "renovate", Name: "renovate"}, Options{MinReleaseAge: time.Hour, MaxPages: maxPages})
			if maxPages == 1 {
				if err != nil || tags == nil || !tags.Truncated || len(tags.Tags) != 1 || requests.Load() != 1 {
					t.Fatalf("expected bounded results without requesting page 2: tags=%+v err=%v requests=%d", tags, err, requests.Load())
				}
				return
			}
			if tags != nil || err == nil || !strings.Contains(err.Error(), message) || !strings.Contains(err.Error(), "HTTP 403") || !strings.Contains(err.Error(), "page=2") || requests.Load() != 2 {
				t.Fatalf("must reject partial results with Hub error details: tags=%+v err=%v requests=%d", tags, err, requests.Load())
			}
		})
	}
}

func TestDockerHubCooldownCandidateFailures(t *testing.T) {
	old := time.Now().Add(-30 * 24 * time.Hour).UTC().Format(time.RFC3339)
	for _, tc := range []struct {
		name        string
		status      int
		body        string
		wantError   string
		wantLookups int
	}{
		{"forbidden", 403, `{"message":"access denied"}`, "access denied", 1},
		{"deleted tag", 404, `{"message":"tag no longer exists"}`, "tag no longer exists", 1},
		{"rate limited", 429, `{"message":"too many requests"}`, "too many requests", 1},
		{"server error", 500, `upstream unavailable`, "upstream unavailable", 1},
		{"malformed JSON", 200, `not JSON`, "failed to decode", 1},
		{"wrong tag", 200, `{"name":"different"}`, "unexpected Docker Hub tag", 1},
		{"missing digest", 200, fmt.Sprintf(`{"name":"TAG","tag_last_pushed":%q}`, old), "no valid digest", 1},
		{"invalid digest", 200, fmt.Sprintf(`{"name":"TAG","tag_last_pushed":%q,"digest":"invalid"}`, old), "no valid digest", 1},
		{"unknown time", 200, `{"name":"TAG"}`, "no eligible", 2},
		{"invalid time", 200, `{"name":"TAG","tag_last_pushed":"invalid"}`, "no eligible", 2},
		{"recent", 200, fmt.Sprintf(`{"name":"TAG","tag_last_pushed":%q}`, time.Now().UTC().Format(time.RFC3339)), "no eligible", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var lookups int
			r := newHubTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
				switch req.URL.Path {
				case "/v2/":
					fmt.Fprint(w, `{}`)
				case "/v2/library/nginx/tags/list":
					fmt.Fprint(w, `{"name":"library/nginx","tags":["1.0.0","2.0.0"]}`)
				case "/v2/repositories/library/nginx/tags/":
					fmt.Fprint(w, `{"results":[]}`)
				case "/v2/repositories/library/nginx/tags/1.0.0/", "/v2/repositories/library/nginx/tags/2.0.0/":
					lookups++
					tag := "2.0.0"
					if strings.HasSuffix(req.URL.Path, "/1.0.0/") {
						tag = "1.0.0"
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tc.status)
					fmt.Fprint(w, strings.ReplaceAll(tc.body, "TAG", tag))
				default:
					t.Errorf("unexpected request: %s", req.URL)
					w.WriteHeader(http.StatusInternalServerError)
				}
			})
			ref := ImageReference{Name: "nginx"}
			opts := Options{MinReleaseAge: 7 * 24 * time.Hour}
			for _, action := range []string{"latest", "show"} {
				lookups = 0
				var err error
				if action == "latest" {
					_, err = r.GetLatestTag(ref, opts)
				} else {
					_, err = r.GetImageInfo(ref, opts)
				}
				if err == nil || !strings.Contains(err.Error(), tc.wantError) || lookups != tc.wantLookups {
					t.Errorf("%s: lookups=%d err=%v", action, lookups, err)
				}
			}
			ref.Tag = "2.0.0"
			if _, err := r.GetImageInfo(ref, opts); err == nil {
				t.Fatal("explicit show must reject an unchecked tag")
			}
		})
	}
}

func TestDockerHubCooldownOCIListFailure(t *testing.T) {
	r := newHubTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		if req.Host == "hub.docker.com" {
			t.Error("must not contact Hub when OCI listing fails")
		}
		if req.URL.Path == "/v2/" {
			fmt.Fprint(w, `{}`)
			return
		}
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"errors":[{"code":"DENIED","message":"access denied"}]}`)
	})
	_, err := r.GetLatestTag(ImageReference{Name: "nginx"}, Options{MinReleaseAge: time.Hour})
	if err == nil || !strings.Contains(err.Error(), "failed to list tags") {
		t.Fatalf("expected OCI listing error, got %v", err)
	}
}

func TestDockerHubVersionsPageBudget(t *testing.T) {
	old := time.Now().Add(-30 * 24 * time.Hour).UTC().Format(time.RFC3339)
	recent := time.Now().UTC().Format(time.RFC3339)
	for _, tc := range []struct {
		name         string
		maxPages     int
		minAge       time.Duration
		wantTags     []string
		wantRequests int32
		truncated    bool
	}{
		{"before age filtering", 1, 3 * 24 * time.Hour, nil, 1, true},
		{"ends exactly at budget", 2, 3 * 24 * time.Hour, []string{"1.0.0"}, 2, false},
		{"ends before budget", 10, 3 * 24 * time.Hour, []string{"1.0.0"}, 2, false},
		{"unlimited default", 0, 3 * 24 * time.Hour, []string{"1.0.0"}, 2, false},
		{"without cooldown", 1, 0, []string{"3.0.0", "2.0.0"}, 1, true},
		{"negative budget", -1, 0, nil, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var requests atomic.Int32
			r := newHubTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
				requests.Add(1)
				if req.URL.Query().Get("page") == "2" {
					fmt.Fprintf(w, `{"results":[{"name":"1.0.0","tag_last_pushed":%q}]}`, old)
					return
				}
				fmt.Fprintf(w, `{"next":"?page=2","results":[{"name":"3.0.0","tag_last_pushed":%q},{"name":"2.0.0"}]}`, recent)
			})
			result, err := (&Client{registry: r}).GetTags("docker.io/library/nginx", Options{MinReleaseAge: tc.minAge, MaxPages: tc.maxPages})
			if tc.maxPages < 0 {
				if err == nil || !strings.Contains(err.Error(), "--max-pages") || requests.Load() != 0 {
					t.Fatalf("expected negative budget rejection before HTTP: err=%v requests=%d", err, requests.Load())
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var names []string
			for _, tag := range result.Tags {
				names = append(names, tag.Name)
			}
			if !reflect.DeepEqual(names, tc.wantTags) || result.Truncated != tc.truncated || requests.Load() != tc.wantRequests {
				t.Fatalf("tags=%v truncated=%v requests=%d", names, result.Truncated, requests.Load())
			}
		})
	}
}
