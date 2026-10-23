package container

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestDockerHubCooldownBatchesCandidates(t *testing.T) {
	old := time.Now().Add(-30 * 24 * time.Hour).UTC().Format(time.RFC3339)
	recent := time.Now().UTC().Format(time.RFC3339)
	var names []string
	for version := 1; version <= 300; version++ {
		names = append(names, fmt.Sprintf("%d.0.0", version))
	}
	var batches, direct atomic.Int32
	r := newHubTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch req.URL.Path {
		case "/v2/":
			fmt.Fprint(w, `{}`)
		case "/v2/library/nginx/tags/list":
			_ = json.NewEncoder(w).Encode(map[string]any{"name": "library/nginx", "tags": names})
		case "/v2/repositories/library/nginx/tags/":
			batches.Add(1)
			first, next := 300, "?page=2"
			if req.URL.Query().Get("page") == "2" {
				first, next = 200, ""
			}
			var results []dockerHubTag
			for version := first; version > first-100; version-- {
				pushed := recent
				if version <= 150 {
					pushed = old
				}
				results = append(results, dockerHubTag{Name: fmt.Sprintf("%d.0.0", version), TagLastPushed: pushed, Digest: testDigest})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"next": next, "results": results})
		default:
			direct.Add(1)
			tag := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/v2/repositories/library/nginx/tags/"), "/")
			version, err := strconv.Atoi(strings.TrimSuffix(tag, ".0.0"))
			if err != nil {
				t.Errorf("unexpected request: %s", req.URL)
			}
			pushed := recent
			if version <= 150 {
				pushed = old
			}
			_ = json.NewEncoder(w).Encode(dockerHubTag{Name: tag, TagLastPushed: pushed, Digest: testDigest})
		}
	})
	tag, err := r.GetLatestTag(ImageReference{Name: "nginx"}, Options{MinReleaseAge: 7 * 24 * time.Hour})
	if err != nil || tag.Name != "150.0.0" || tag.Digest != testDigest {
		t.Fatalf("latest = %+v, err = %v", tag, err)
	}
	if batches.Load() != 2 || direct.Load() != 0 {
		t.Fatalf("expected 2 batch requests and no direct lookups; batches=%d direct=%d", batches.Load(), direct.Load())
	}
}

func TestDockerHubCooldownBatchPreservesSemverSelection(t *testing.T) {
	const higherDigest = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	old := time.Now().Add(-30 * 24 * time.Hour).UTC().Format(time.RFC3339)
	recent := time.Now().UTC().Format(time.RFC3339)
	for _, tc := range []struct {
		name       string
		cachedHigh bool
		pushedHigh string
		wantTag    string
		wantDigest string
		wantDirect int32
	}{
		{"missing higher candidate", false, old, "9.0.0", higherDigest, 1},
		{"missing recent candidate", false, recent, "2.0.0", testDigest, 1},
		{"cached recent candidate", true, recent, "2.0.0", testDigest, 0},
		{"cached invalid timestamp", true, "invalid", "2.0.0", testDigest, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var direct atomic.Int32
			r := newHubTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch req.URL.Path {
				case "/v2/":
					fmt.Fprint(w, `{}`)
				case "/v2/library/nginx/tags/list":
					fmt.Fprint(w, `{"tags":["1.0.0","2.0.0","9.0.0"]}`)
				case "/v2/repositories/library/nginx/tags/":
					results := []dockerHubTag{{Name: "2.0.0", TagLastPushed: old, Digest: testDigest}}
					if tc.cachedHigh {
						results = append(results, dockerHubTag{Name: "9.0.0", TagLastPushed: tc.pushedHigh, Digest: higherDigest})
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"results": results})
				case "/v2/repositories/library/nginx/tags/9.0.0/":
					direct.Add(1)
					_ = json.NewEncoder(w).Encode(dockerHubTag{Name: "9.0.0", TagLastPushed: tc.pushedHigh, Digest: higherDigest})
				default:
					t.Errorf("unnecessary request: %s", req.URL)
					w.WriteHeader(http.StatusNotFound)
				}
			})
			tag, err := r.GetLatestTag(ImageReference{Name: "nginx"}, Options{MinReleaseAge: 7 * 24 * time.Hour})
			if err != nil || tag.Name != tc.wantTag || tag.Digest != tc.wantDigest || direct.Load() != tc.wantDirect {
				t.Fatalf("latest=%+v err=%v direct=%d", tag, err, direct.Load())
			}
		})
	}
}

func TestDockerHubCooldownBatchBoundAndDuplicates(t *testing.T) {
	old := time.Now().Add(-30 * 24 * time.Hour).UTC().Format(time.RFC3339)
	recent := time.Now().UTC().Format(time.RFC3339)
	var batches, direct atomic.Int32
	r := newHubTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch req.URL.Path {
		case "/v2/":
			fmt.Fprint(w, `{}`)
		case "/v2/library/nginx/tags/list":
			fmt.Fprint(w, `{"tags":["1.0.0","9.0.0"]}`)
		case "/v2/repositories/library/nginx/tags/":
			page := batches.Add(1)
			if page > 3 {
				t.Error("prefetch must not enumerate the entire repository")
				w.WriteHeader(http.StatusForbidden)
				return
			}
			pushed := old
			if page == 1 {
				pushed = recent
			}
			// A tag can move between pages. Do not overwrite the first snapshot
			// with an older duplicate from a later page.
			_ = json.NewEncoder(w).Encode(map[string]any{
				"next":    fmt.Sprintf("?page=%d", page+1),
				"results": []dockerHubTag{{Name: "9.0.0", TagLastPushed: pushed, Digest: testDigest}},
			})
		case "/v2/repositories/library/nginx/tags/1.0.0/":
			direct.Add(1)
			_ = json.NewEncoder(w).Encode(dockerHubTag{Name: "1.0.0", TagLastPushed: old, Digest: testDigest})
		default:
			t.Errorf("unexpected request: %s", req.URL)
			w.WriteHeader(http.StatusNotFound)
		}
	})
	tag, err := r.GetLatestTag(ImageReference{Name: "nginx"}, Options{MinReleaseAge: 7 * 24 * time.Hour})
	if err != nil || tag.Name != "1.0.0" || batches.Load() != 3 || direct.Load() != 1 {
		t.Fatalf("latest=%+v err=%v batches=%d direct=%d", tag, err, batches.Load(), direct.Load())
	}
}

func TestDockerHubCooldownBatchFailureFallsBack(t *testing.T) {
	old := time.Now().Add(-30 * 24 * time.Hour).UTC().Format(time.RFC3339)
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"anonymous limit", 403, `{"message":"pagination offset too large for anonymous requests"}`},
		{"rate limited", 429, `{"message":"too many requests"}`},
		{"invalid JSON", 200, `invalid JSON`},
		{"cross origin", 200, `{"next":"https://attacker.example/tags","results":[]}`},
		{"loop", 200, `{"next":"?page_size=100&ordering=last_updated","results":[]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var batches, direct atomic.Int32
			r := newHubTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch req.URL.Path {
				case "/v2/":
					fmt.Fprint(w, `{}`)
				case "/v2/library/nginx/tags/list":
					fmt.Fprint(w, `{"tags":["1.0.0"]}`)
				case "/v2/repositories/library/nginx/tags/":
					batches.Add(1)
					w.WriteHeader(tc.status)
					fmt.Fprint(w, tc.body)
				case "/v2/repositories/library/nginx/tags/1.0.0/":
					direct.Add(1)
					_ = json.NewEncoder(w).Encode(dockerHubTag{Name: "1.0.0", TagLastPushed: old, Digest: testDigest})
				default:
					t.Errorf("unexpected request: %s", req.URL)
					w.WriteHeader(http.StatusNotFound)
				}
			})
			tag, err := r.GetLatestTag(ImageReference{Name: "nginx"}, Options{MinReleaseAge: time.Hour})
			if err != nil || tag.Name != "1.0.0" || batches.Load() != 1 || direct.Load() != 1 {
				t.Fatalf("latest=%+v err=%v batches=%d direct=%d", tag, err, batches.Load(), direct.Load())
			}
		})
	}
}

func TestDockerHubCooldownBatchPartialFailureAndFreshness(t *testing.T) {
	old := time.Now().Add(-30 * 24 * time.Hour).UTC().Format(time.RFC3339)
	recent := time.Now().UTC().Format(time.RFC3339)
	var batches, direct atomic.Int32
	r := newHubTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch req.URL.Path {
		case "/v2/":
			fmt.Fprint(w, `{}`)
		case "/v2/library/nginx/tags/list":
			fmt.Fprint(w, `{"tags":["1.0.0","9.0.0"]}`)
		case "/v2/repositories/library/nginx/tags/":
			if req.URL.Query().Get("page") == "2" {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprint(w, `{"message":"sign in to page further"}`)
				return
			}
			query := batches.Add(1)
			pushed := recent
			if query == 2 {
				pushed = old
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"next":    "?page=2",
				"results": []dockerHubTag{{Name: "9.0.0", TagLastPushed: pushed, Digest: testDigest}},
			})
		case "/v2/repositories/library/nginx/tags/1.0.0/":
			direct.Add(1)
			_ = json.NewEncoder(w).Encode(dockerHubTag{Name: "1.0.0", TagLastPushed: old, Digest: testDigest})
		default:
			t.Errorf("unnecessary request: %s", req.URL)
			w.WriteHeader(http.StatusNotFound)
		}
	})
	for _, want := range []string{"1.0.0", "9.0.0"} {
		tag, err := r.GetLatestTag(ImageReference{Name: "nginx"}, Options{MinReleaseAge: time.Hour})
		if err != nil || tag.Name != want {
			t.Fatalf("latest=%+v err=%v, want %s", tag, err, want)
		}
	}
	if batches.Load() != 2 || direct.Load() != 1 {
		t.Fatalf("reuse successful pages within a query, not between queries: batches=%d direct=%d", batches.Load(), direct.Load())
	}
}

func TestDockerHubCooldownBatchRejectsInvalidDigest(t *testing.T) {
	old := time.Now().Add(-30 * 24 * time.Hour).UTC().Format(time.RFC3339)
	r := newHubTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/v2/":
			fmt.Fprint(w, `{}`)
		case "/v2/library/nginx/tags/list":
			fmt.Fprint(w, `{"tags":["1.0.0","9.0.0"]}`)
		case "/v2/repositories/library/nginx/tags/":
			fmt.Fprintf(w, `{"results":[{"name":"9.0.0","tag_last_pushed":%q,"digest":"invalid"}]}`, old)
		default:
			t.Errorf("must not bypass an invalid selected digest: %s", req.URL)
			w.WriteHeader(http.StatusNotFound)
		}
	})
	_, err := r.GetLatestTag(ImageReference{Name: "nginx"}, Options{MinReleaseAge: time.Hour})
	if err == nil || !strings.Contains(err.Error(), "no valid digest") {
		t.Fatalf("invalid cached digest accepted: %v", err)
	}
}
