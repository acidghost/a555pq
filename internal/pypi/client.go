package pypi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/acidghost/a555pq/internal/formatter"
)

const pypiAPIURL = "https://pypi.org/pypi"

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
	url := fmt.Sprintf("%s/%s/json", pypiAPIURL, name)

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
	for version, files := range info.Releases {
		if len(files) > 0 {
			versions = append(versions, formatter.VersionItem{
				Version:    version,
				UploadDate: files[0].UploadTime.Format(time.RFC3339),
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

	return info.Info.Version, nil
}
