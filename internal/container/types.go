package container

type ImageReference struct {
	Registry     string
	Organization string
	Name         string
	Tag          string
}

type ImageInfo struct {
	Name         string
	Description  string
	LatestTag    string
	TagDate      string
	Manifest     *ManifestInfo
	Size         string
	Registry     string
	FullImageRef string
}

type ManifestInfo struct {
	Digest       string
	MediaType    string
	Architecture string
	OS           string
	Layers       []string
}

type TagInfo struct {
	Name      string
	CreatedAt string
	Digest    string
	Size      string
}

type TagMetadata struct {
	Size   string
	Date   string
	Digest string
}
