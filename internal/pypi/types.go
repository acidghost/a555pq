package pypi

import (
	"encoding/json"
	"strings"
	"time"
)

type PackageInfo struct {
	Info        Info              `json:"info"`
	Releases    map[string][]File `json:"releases"`
	LastUpdated int64             `json:"last_serial"`
}

type Info struct {
	Author             string            `json:"author"`
	AuthorEmail        string            `json:"author_email"`
	BugTrackURL        string            `json:"bugtrack_url"`
	Classifiers        []string          `json:"classifiers"`
	Description        string            `json:"description"`
	DescriptionContent string            `json:"description_content_type"`
	DocsURL            string            `json:"docs_url"`
	DownloadURL        string            `json:"download_url"`
	HomePage           string            `json:"home_page"`
	Keywords           Keywords          `json:"keywords"`
	License            string            `json:"license"`
	LicenseFiles       []string          `json:"license_files"`
	Maintainer         string            `json:"maintainer"`
	MaintainerEmail    string            `json:"maintainer_email"`
	Name               string            `json:"name"`
	PackageURL         string            `json:"package_url"`
	ProjectURL         string            `json:"project_url"`
	ProjectURLs        map[string]string `json:"project_urls"`
	ReleaseURL         string            `json:"release_url"`
	RequiresDist       []string          `json:"requires_dist"`
	RequiresPython     string            `json:"requires_python"`
	Summary            string            `json:"summary"`
	Version            string            `json:"version"`
	Yanked             bool              `json:"yanked"`
	YankedReason       string            `json:"yanked_reason"`
}

type Keywords []string

func (k *Keywords) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*k = strings.Fields(s)
		return nil
	}
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	*k = arr
	return nil
}

func (k Keywords) Slice() []string {
	return k
}

type File struct {
	PackageType   string            `json:"packagetype"`
	PythonVersion string            `json:"python_version"`
	UploadTime    time.Time         `json:"upload_time_iso_8601"`
	URL           string            `json:"url"`
	Size          int64             `json:"size"`
	Filename      string            `json:"filename"`
	Hashes        map[string]string `json:"hashes"`
}
