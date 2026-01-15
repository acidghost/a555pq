package npm

import (
	"encoding/json"

	"github.com/acidghost/a555pq/internal/formatter"
)

type PackageInfo struct {
	ID         string                 `json:"_id"`
	Name       string                 `json:"name"`
	Revision   string                 `json:"_rev"`
	DistTags   map[string]string      `json:"dist-tags"`
	Versions   map[string]VersionInfo `json:"versions"`
	Time       map[string]string      `json:"time"`
	Author     FlexibleAuthor         `json:"author"`
	Repository Repository             `json:"repository"`
	Readme     string                 `json:"readme"`
}

type VersionInfo struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Homepage        string            `json:"homepage"`
	Repository      Repository        `json:"repository"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Author          FlexibleAuthor    `json:"author"`
	License         string            `json:"license"`
	Readme          string            `json:"readme"`
	ReadmeFilename  string            `json:"readmeFilename"`
	Dist            Dist              `json:"dist"`
	Maintainers     []FlexibleAuthor  `json:"maintainers"`
}

type FlexibleAuthor struct {
	Name  string
	Email string
	URL   string
}

func (fa *FlexibleAuthor) UnmarshalJSON(data []byte) error {
	var author formatter.Author
	if err := json.Unmarshal(data, &author); err == nil {
		fa.Name = author.Name
		fa.Email = author.Email
		fa.URL = author.URL
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		fa.Name = s
		return nil
	}
	return json.Unmarshal(data, &author)
}

func (fa *FlexibleAuthor) ToAuthor() formatter.Author {
	return formatter.Author{
		Name:  fa.Name,
		Email: fa.Email,
		URL:   fa.URL,
	}
}

type Repository struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type Dist struct {
	Shasum  string `json:"shasum"`
	Tarball string `json:"tarball"`
}

func (v VersionInfo) MarshalJSON() ([]byte, error) {
	type Alias VersionInfo
	return json.Marshal(&struct {
		*Alias
		Maintainers []formatter.Author `json:"maintainers,omitempty"`
	}{
		Alias:       (*Alias)(&v),
		Maintainers: nil,
	})
}
