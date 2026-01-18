package github

import "time"

type Repository struct {
	ID               int64     `json:"id"`
	NodeID           string    `json:"node_id"`
	Name             string    `json:"name"`
	FullName         string    `json:"full_name"`
	Owner            Owner     `json:"owner"`
	Private          bool      `json:"private"`
	Description      string    `json:"description"`
	Fork             bool      `json:"fork"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	PushedAt         time.Time `json:"pushed_at"`
	Homepage         string    `json:"homepage"`
	Size             int       `json:"size"`
	StargazersCount  int       `json:"stargazers_count"`
	WatchersCount    int       `json:"watchers_count"`
	Language         string    `json:"language"`
	HasIssues        bool      `json:"has_issues"`
	HasProjects      bool      `json:"has_projects"`
	HasDownloads     bool      `json:"has_downloads"`
	HasWiki          bool      `json:"has_wiki"`
	HasPages         bool      `json:"has_pages"`
	ForksCount       int       `json:"forks_count"`
	OpenIssuesCount  int       `json:"open_issues_count"`
	DefaultBranch    string    `json:"default_branch"`
	License          *License  `json:"license"`
	Archived         bool      `json:"archived"`
	Disabled         bool      `json:"disabled"`
	Forks            int       `json:"forks"`
	OpenIssues       int       `json:"open_issues"`
	Watchers         int       `json:"watchers"`
	SubscribersCount int       `json:"subscribers_count"`
}

type Owner struct {
	Login      string `json:"login"`
	ID         int64  `json:"id"`
	NodeID     string `json:"node_id"`
	AvatarURL  string `json:"avatar_url"`
	GravatarID string `json:"gravatar_id"`
	URL        string `json:"url"`
	HTMLURL    string `json:"html_url"`
	Type       string `json:"type"`
	SiteAdmin  bool   `json:"site_admin"`
}

type License struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	SPDXID string `json:"spdx_id"`
	URL    string `json:"url"`
	NodeID string `json:"node_id"`
}

type Release struct {
	ID              int64     `json:"id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     string    `json:"published_at"`
	Assets          []Asset   `json:"assets"`
	Author          Owner     `json:"author"`
	HTMLURL         string    `json:"html_url"`
	URL             string    `json:"url"`
	AssetsURL       string    `json:"assets_url"`
	UploadURL       string    `json:"upload_url"`
	TarballURL      string    `json:"tarball_url"`
	ZipballURL      string    `json:"zipball_url"`
	Body            string    `json:"body"`
}

type Asset struct {
	URL                string    `json:"url"`
	ID                 int64     `json:"id"`
	NodeID             string    `json:"node_id"`
	Name               string    `json:"name"`
	Label              string    `json:"label"`
	Uploader           Owner     `json:"uploader"`
	ContentType        string    `json:"content_type"`
	State              string    `json:"state"`
	Size               int64     `json:"size"`
	DownloadCount      int       `json:"download_count"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	BrowserDownloadURL string    `json:"browser_download_url"`
}

type Tag struct {
	Name       string    `json:"name"`
	ZipballURL string    `json:"zipball_url"`
	TarballURL string    `json:"tarball_url"`
	Commit     TagCommit `json:"commit"`
	NodeID     string    `json:"node_id"`
}

type TagCommit struct {
	SHA  string `json:"sha"`
	URL  string `json:"url"`
	Date string `json:"date"`
}

type GraphQLResponse struct {
	Data   *GraphQLData   `json:"data"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

type GraphQLData struct {
	Repository *GraphQLRepository `json:"repository"`
}

type GraphQLRepository struct {
	Releases *GraphQLConnection `json:"releases"`
	Refs     *GraphQLConnection `json:"refs"`
}

type GraphQLConnection struct {
	Nodes []any `json:"nodes"`
}

type GraphQLReleaseNode struct {
	TagName      string `json:"tagName"`
	IsPrerelease bool   `json:"isPrerelease"`
	CreatedAt    string `json:"createdAt"`
	PublishedAt  string `json:"publishedAt"`
}

type GraphQLRefNode struct {
	Name   string            `json:"name"`
	Target *GraphQLRefTarget `json:"target"`
}

type GraphQLRefTarget struct {
	Author    *GraphQLCommitAuthor `json:"author"`
	Committer *GraphQLCommitAuthor `json:"committer"`
	Tagger    *GraphQLTagger       `json:"tagger"`
	Target    *GraphQLRefTarget    `json:"target"`
}

type GraphQLCommitAuthor struct {
	Date string `json:"date"`
}

type GraphQLTagger struct {
	Date string `json:"date"`
}

type GraphQLError struct {
	Message string   `json:"message"`
	Type    string   `json:"type"`
	Path    []string `json:"path"`
}
