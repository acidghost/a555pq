package container

import "time"

// Options controls optional container query behavior.
type Options struct {
	// MinReleaseAge requires a known Docker Hub tag_last_pushed timestamp at
	// least this old. Zero disables the cooldown.
	MinReleaseAge time.Duration

	// MaxPages bounds Docker Hub tag-list pages before age filtering.
	// Zero leaves retrieval unlimited.
	MaxPages int
}

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

// TagList tracks completeness separately from the age-filtered tags.
type TagList struct {
	Tags      []TagInfo
	Truncated bool
}

type TagMetadata struct {
	Size   string
	Date   string
	Digest string
}
