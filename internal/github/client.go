package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/acidghost/a555pq/internal/formatter"
)

const githubAPIURL = "https://api.github.com"
const graphqlAPIURL = "https://api.github.com/graphql"

type Client struct {
	httpClient *http.Client
	token      string
	forceREST  bool
}

func getGitHubToken() string {
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token
	}

	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func NewClient(forceREST bool) *Client {
	token := getGitHubToken()
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		token:     token,
		forceREST: forceREST,
	}
}

func (c *Client) GetPackageInfo(name string) (*Repository, error) {
	url := fmt.Sprintf("%s/repos/%s", githubAPIURL, name)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("repository '%s' not found", name)
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("rate limit exceeded, consider setting GITHUB_TOKEN environment variable")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var repo Repository
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &repo, nil
}

func (c *Client) GetVersions(name string) ([]formatter.VersionItem, error) {
	if c.token != "" && !c.forceREST {
		return c.getVersionsGraphQL(name)
	}
	return c.getVersionsREST(name)
}

func (c *Client) getVersionsREST(name string) ([]formatter.VersionItem, error) {
	var (
		releaseURL = fmt.Sprintf("%s/repos/%s/releases", githubAPIURL, name)
		tagURL     = fmt.Sprintf("%s/repos/%s/tags", githubAPIURL, name)

		releases []Release
		tags     []Tag
	)

	if err := c.fetchJSON(releaseURL, &releases); err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}

	if err := c.fetchJSON(tagURL, &tags); err != nil {
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
	}

	releaseTags := make(map[string]struct{})
	var versions []formatter.VersionItem

	for _, release := range releases {
		releaseTags[release.TagName] = struct{}{}
	}

	for _, release := range releases {
		if !release.Prerelease {
			versions = append(versions, formatter.VersionItem{
				Version:    release.TagName,
				UploadDate: release.PublishedAt,
			})
		}
	}

	for _, tag := range tags {
		if _, ok := releaseTags[tag.Name]; !ok {
			versions = append(versions, formatter.VersionItem{
				Version:    tag.Name,
				UploadDate: "",
			})
		}
	}

	return versions, nil
}

func (c *Client) getVersionsGraphQL(name string) ([]formatter.VersionItem, error) {
	owner, repoName, err := parseOwnerRepo(name)
	if err != nil {
		return nil, err
	}

	query := `
		query($owner: String!, $name: String!) {
			repository(owner: $owner, name: $name) {
				releases(first: 100, orderBy: {field: CREATED_AT, direction: DESC}) {
					nodes {
						tagName
						isPrerelease
						createdAt
						publishedAt
					}
				}
				refs(refPrefix: "refs/tags/", first: 100, orderBy: {field: TAG_COMMIT_DATE, direction: DESC}) {
					nodes {
						name
						target {
							... on Commit {
								author { date }
								committer { date }
							}
							... on Tag {
								tagger { date }
								target {
									... on Commit {
										author { date }
									}
								}
							}
						}
					}
				}
			}
		}
	`

	variables := map[string]any{
		"owner": owner,
		"name":  repoName,
	}

	response, err := c.executeGraphQLQuery(query, variables)
	if err != nil {
		return nil, err
	}

	if len(response.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %v", response.Errors[0].Message)
	}

	if response.Data == nil || response.Data.Repository == nil {
		return nil, fmt.Errorf("repository '%s' not found", name)
	}

	releaseTags := make(map[string]string)
	var versions []formatter.VersionItem

	if response.Data.Repository.Releases != nil && len(response.Data.Repository.Releases.Nodes) > 0 {
		for _, node := range response.Data.Repository.Releases.Nodes {
			release, ok := node.(map[string]any)
			if !ok {
				continue
			}

			tagName, _ := release["tagName"].(string)
			isPrerelease, _ := release["isPrerelease"].(bool)
			publishedAt, _ := release["publishedAt"].(string)

			if !isPrerelease && tagName != "" {
				versions = append(versions, formatter.VersionItem{
					Version:    tagName,
					UploadDate: publishedAt,
				})
				releaseTags[tagName] = publishedAt
			}
		}
	}

	if response.Data.Repository.Refs != nil {
		for _, node := range response.Data.Repository.Refs.Nodes {
			ref, ok := node.(map[string]any)
			if !ok {
				continue
			}

			name, _ := ref["name"].(string)
			if _, exists := releaseTags[name]; exists {
				continue
			}

			var date string
			if target, ok := ref["target"].(map[string]any); ok {
				if tagger, ok := target["tagger"].(map[string]any); ok {
					if d, ok := tagger["date"].(string); ok && d != "" {
						date = d
					}
				}
				if date == "" {
					if author, ok := target["author"].(map[string]any); ok {
						if d, ok := author["date"].(string); ok {
							date = d
						}
					}
				}
				if date == "" {
					if committer, ok := target["committer"].(map[string]any); ok {
						if d, ok := committer["date"].(string); ok {
							date = d
						}
					}
				}
			}

			if name != "" {
				versions = append(versions, formatter.VersionItem{
					Version:    name,
					UploadDate: date,
				})
			}
		}
	}

	if len(versions) == 0 {
		return []formatter.VersionItem{}, nil
	}

	return versions, nil
}

func (c *Client) executeGraphQLQuery(query string, variables map[string]any) (*GraphQLResponse, error) {
	requestBody := map[string]any{
		"query":     query,
		"variables": variables,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", graphqlAPIURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute GraphQL query: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication failed, check GITHUB_TOKEN")
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

func parseOwnerRepo(name string) (string, string, error) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository format: expected 'owner/repo', got '%s'", name)
	}
	return parts[0], parts[1], nil
}

func (c *Client) GetLatestVersion(name string) (string, error) {
	latestURL := fmt.Sprintf("%s/repos/%s/releases/latest", githubAPIURL, name)

	req, err := http.NewRequest("GET", latestURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var release Release
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return "", fmt.Errorf("failed to decode response: %w", err)
		}
		if !release.Prerelease {
			return release.TagName, nil
		}
	}

	versions, err := c.GetVersions(name)
	if err != nil {
		return "", err
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no versions found for repository '%s'", name)
	}

	return versions[0].Version, nil
}

func (c *Client) fetchJSON(url string, v any) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("rate limit exceeded for unauthenticated requests. Set GITHUB_TOKEN environment variable to use GraphQL API with higher rate limits")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
