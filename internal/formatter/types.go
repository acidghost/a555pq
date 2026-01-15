package formatter

type Author struct {
	Name  string
	Email string
	URL   string
}

type VersionItem struct {
	Version    string
	UploadDate string
}

type ShowOutput struct {
	Name         string
	Version      string
	Description  string
	Author       string
	AuthorEmail  string
	License      string
	HomePage     string
	Dependencies []string
}

type VersionsOutput struct {
	Package  string
	Versions []VersionItem
}

type LatestOutput struct {
	Package string
	Version string
}

type BrowseOutput struct {
	Package string
	URL     string
	Opened  bool
}
