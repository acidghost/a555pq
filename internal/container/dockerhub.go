package container

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

const dockerHubOrigin = "https://hub.docker.com"

type dockerHubTag struct {
	Name          string `json:"name"`
	TagLastPushed string `json:"tag_last_pushed"`
	Digest        string `json:"digest"`
	FullSize      int64  `json:"full_size"`
}

func (tag dockerHubTag) info() TagInfo {
	var size string
	if tag.FullSize > 0 {
		size = formatBytes(tag.FullSize)
	}
	return TagInfo{
		Name:      tag.Name,
		CreatedAt: tag.TagLastPushed,
		Digest:    tag.Digest,
		Size:      size,
	}
}

func (r *UnifiedRegistry) dockerHubRepository(ref ImageReference) (name.Repository, error) {
	repo, err := r.parseRef(ref)
	if err != nil {
		return name.Repository{}, fmt.Errorf("failed to parse reference: %w", err)
	}
	if !r.isDockerHub(repo) {
		return name.Repository{}, fmt.Errorf("cooldown and page-limited queries are supported only for Docker Hub images")
	}
	repository := repo.RepositoryStr()
	if !strings.Contains(repository, "/") {
		repository = "library/" + repository
	}
	// Normalize Hub aliases for OCI requests as well as Hub API requests.
	return name.NewRepository(name.DefaultRegistry + "/" + repository)
}

func dockerHubTagsURL(repo name.Repository) string {
	return dockerHubOrigin + "/v2/repositories/" + repo.RepositoryStr() + "/tags/"
}

type dockerHubTagsPage struct {
	Next    string         `json:"next"`
	Results []dockerHubTag `json:"results"`
}

func (r *UnifiedRegistry) fetchDockerHubTagsPage(pageURL string) (*dockerHubTagsPage, error) {
	var page dockerHubTagsPage
	if err := r.fetchDockerHubJSON(pageURL, &page); err != nil {
		return nil, err
	}
	if page.Next != "" {
		base, err := url.Parse(pageURL)
		if err != nil {
			return nil, err
		}
		next, err := base.Parse(page.Next)
		if err != nil || !isDockerHubURL(next) {
			return nil, fmt.Errorf("invalid or cross-origin Docker Hub pagination link: %q", page.Next)
		}
		page.Next = next.String()
	}
	return &page, nil
}

// fetchDockerHubTags reads registry push times, never publisher-controlled image
// creation times. Like Renovate, use tag_last_pushed rather than last_updated:
// metadata edits must not be mistaken for a new image push.
func (r *UnifiedRegistry) fetchDockerHubTags(repo name.Repository, maxPages int) (*TagList, error) {
	pageURL := dockerHubTagsURL(repo) + "?page_size=100&ordering=last_updated"
	seen := make(map[string]bool)
	result := &TagList{}
	for pageURL != "" {
		if seen[pageURL] {
			return nil, fmt.Errorf("invalid Docker Hub pagination: loop detected")
		}
		if maxPages > 0 && len(seen) >= maxPages {
			result.Truncated = true
			break
		}
		if len(seen) >= 1000 {
			return nil, fmt.Errorf("invalid Docker Hub pagination: page limit exceeded")
		}
		seen[pageURL] = true
		page, err := r.fetchDockerHubTagsPage(pageURL)
		if err != nil {
			return nil, err
		}
		for _, tag := range page.Results {
			result.Tags = append(result.Tags, tag.info())
		}
		pageURL = page.Next
	}
	return result, nil
}

// Prefetch is only an optimization, never a complete list of available tags.
// A small bound avoids the anonymous pagination limit and excessive requests.
const dockerHubPrefetchPages = 3

func (r *UnifiedRegistry) prefetchDockerHubTags(repo name.Repository) (map[string]TagInfo, error) {
	pageURL := dockerHubTagsURL(repo) + "?page_size=100&ordering=last_updated"
	seen := make(map[string]bool)
	metadata := make(map[string]TagInfo)
	for pages := 0; pageURL != "" && pages < dockerHubPrefetchPages; pages++ {
		if seen[pageURL] {
			return metadata, fmt.Errorf("invalid Docker Hub pagination: loop detected")
		}
		seen[pageURL] = true
		page, err := r.fetchDockerHubTagsPage(pageURL)
		if err != nil {
			return metadata, err
		}
		for _, tag := range page.Results {
			// Keep the first snapshot if a moving tag appears on multiple pages.
			if _, exists := metadata[tag.Name]; !exists {
				metadata[tag.Name] = tag.info()
			}
		}
		pageURL = page.Next
	}
	return metadata, nil
}

func (r *UnifiedRegistry) fetchDockerHubTag(repo name.Repository, tag string) (TagInfo, error) {
	var result dockerHubTag
	if err := r.fetchDockerHubJSON(dockerHubTagsURL(repo)+url.PathEscape(tag)+"/", &result); err != nil {
		return TagInfo{}, err
	}
	if result.Name != tag {
		return TagInfo{}, fmt.Errorf("unexpected Docker Hub tag %q, requested %q", result.Name, tag)
	}
	return result.info(), nil
}

func isDockerHubURL(u *url.URL) bool {
	return u != nil && u.Scheme == "https" && u.Host == "hub.docker.com" && u.User == nil
}

func (r *UnifiedRegistry) fetchDockerHubJSON(requestURL string, target any) error {
	// Pagination and redirects must not send requests to an arbitrary host.
	client := *r.httpClient
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if !isDockerHubURL(req.URL) || len(via) >= 10 {
			return fmt.Errorf("invalid Docker Hub redirect")
		}
		return nil
	}
	resp, err := client.Get(requestURL)
	if err != nil {
		return fmt.Errorf("failed to fetch Docker Hub URL %q: %w", requestURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		var detail struct {
			Message string `json:"message"`
		}
		message := strings.TrimSpace(string(body))
		if json.Unmarshal(body, &detail) == nil && detail.Message != "" {
			message = detail.Message
		}
		return fmt.Errorf("failed to fetch Docker Hub URL %q: HTTP %d: %q", requestURL, resp.StatusCode, message)
	}
	const maxBodySize = 8 << 20
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize+1))
	if err != nil {
		return fmt.Errorf("failed to read Docker Hub URL %q: %w", requestURL, err)
	}
	if len(body) > maxBodySize {
		return fmt.Errorf("response too large from Docker Hub URL %q", requestURL)
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to decode Docker Hub URL %q: %w", requestURL, err)
	}
	return nil
}

func tagPushedBefore(tag TagInfo, cutoff time.Time) bool {
	pushed, err := time.Parse(time.RFC3339Nano, tag.CreatedAt)
	return err == nil && !pushed.IsZero() && !pushed.After(cutoff)
}

// Unknown or invalid push times are ineligible when a cooldown is requested.
func filterTagsByMinReleaseAge(tags []TagInfo, minAge time.Duration, now time.Time) []TagInfo {
	cutoff := now.Add(-minAge)
	filtered := make([]TagInfo, 0, len(tags))
	for _, tag := range tags {
		if tagPushedBefore(tag, cutoff) {
			filtered = append(filtered, tag)
		}
	}
	return filtered
}

func validateTagDigest(tag TagInfo) error {
	if _, err := v1.NewHash(tag.Digest); err != nil {
		return fmt.Errorf("no valid digest for Docker Hub tag %q: %w", tag.Name, err)
	}
	return nil
}

func (r *UnifiedRegistry) getLatestCooledTag(ref ImageReference, opts Options) (TagInfo, error) {
	repo, err := r.dockerHubRepository(ref)
	if err != nil {
		return TagInfo{}, err
	}
	// Always rank the complete OCI list; Hub's push-time ordering does not tell
	// us which tag has the highest semantic version.
	names, err := remote.List(repo, r.options...)
	if err != nil {
		return TagInfo{}, fmt.Errorf("failed to list tags: %w", err)
	}
	tags := make([]TagInfo, 0, len(names))
	for _, tag := range names {
		tags = append(tags, TagInfo{Name: tag})
	}
	candidates := filterSemverTags(tags)
	if len(candidates) == 0 {
		return TagInfo{}, fmt.Errorf("no semantic version tags found for image '%s'", ref.Name)
	}
	cutoff := time.Now().Add(-opts.MinReleaseAge)
	// Prefetch is optional: successful pages are still useful if a later page
	// fails. Every missing candidate gets a direct age/digest check below.
	metadata, _ := r.prefetchDockerHubTags(repo)
	for _, candidate := range candidates {
		tag, cached := metadata[candidate.Name]
		if !cached {
			tag, err = r.fetchDockerHubTag(repo, candidate.Name)
			if err != nil {
				return TagInfo{}, err
			}
		}
		if !tagPushedBefore(tag, cutoff) {
			continue
		}
		if err := validateTagDigest(tag); err != nil {
			return TagInfo{}, err
		}
		return tag, nil
	}
	return TagInfo{}, fmt.Errorf("no eligible semantic version tags found for image '%s'", ref.Name)
}

func (r *UnifiedRegistry) getCooledImageInfo(ref ImageReference, opts Options) (*ImageInfo, error) {
	var selected TagInfo
	var err error
	if ref.Tag == "" {
		selected, err = r.GetLatestTag(ref, opts)
		if err != nil {
			return nil, err
		}
	} else {
		repo, err := r.dockerHubRepository(ref)
		if err != nil {
			return nil, err
		}
		selected, err = r.fetchDockerHubTag(repo, ref.Tag)
		if err != nil {
			return nil, err
		}
		if !tagPushedBefore(selected, time.Now().Add(-opts.MinReleaseAge)) {
			return nil, fmt.Errorf("tag %q is too recent or has no known push time", ref.Tag)
		}
		if err := validateTagDigest(selected); err != nil {
			return nil, err
		}
	}
	// Use one Hub snapshot for both the age check and the digest. Resolving the
	// mutable tag again could return a different, freshly pushed image.
	return &ImageInfo{
		Name:         ref.Name,
		Description:  r.fetchDockerHubDescription(ref),
		LatestTag:    selected.Name,
		TagDate:      selected.CreatedAt,
		Size:         selected.Size,
		Manifest:     &ManifestInfo{Digest: selected.Digest},
		Registry:     r.getRegistryName(ref.Registry),
		FullImageRef: r.buildFullImageRef(ref) + ":" + selected.Name + "@" + selected.Digest,
	}, nil
}
