package npm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/acidghost/a555pq/internal/formatter"
)

const npmRegistryURL = "https://registry.npmjs.org"

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) GetPackageInfo(name string) (*PackageInfo, error) {
	url := fmt.Sprintf("%s/%s", npmRegistryURL, name)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("package '%s' not found", name)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var info PackageInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &info, nil
}

func (c *Client) GetVersions(name string) ([]formatter.VersionItem, error) {
	info, err := c.GetPackageInfo(name)
	if err != nil {
		return nil, err
	}

	var versions []formatter.VersionItem
	for version := range info.Versions {
		if date, ok := info.Time[version]; ok {
			versions = append(versions, formatter.VersionItem{
				Version:    version,
				UploadDate: date,
			})
		}
	}

	return versions, nil
}

func (c *Client) GetLatestVersion(name string) (string, error) {
	info, err := c.GetPackageInfo(name)
	if err != nil {
		return "", err
	}

	if latest, ok := info.DistTags["latest"]; ok {
		return latest, nil
	}

	return "", fmt.Errorf("no latest version found for package '%s'", name)
}
