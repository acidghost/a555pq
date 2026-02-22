package container

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

type UnifiedRegistry struct {
	options    []remote.Option
	httpClient *http.Client
	cache      map[string]*remote.Descriptor
}

func NewUnifiedRegistry() *UnifiedRegistry {
	return &UnifiedRegistry{
		options: []remote.Option{},
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache: make(map[string]*remote.Descriptor),
	}
}

func (r *UnifiedRegistry) GetImageInfo(ref ImageReference) (*ImageInfo, error) {
	repoRef, err := r.parseRef(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reference: %w", err)
	}

	var targetTag string

	if ref.Tag != "" {
		targetTag = ref.Tag
	} else {
		targetTag = "latest"
	}

	metadata, err := r.fetchTagMetadata(repoRef, targetTag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	description := r.fetchDescription(ref, repoRef.Name(), targetTag)

	fullRef := r.buildFullImageRef(ref)
	if ref.Tag != "" {
		fullRef = fmt.Sprintf("%s:%s", fullRef, ref.Tag)
	} else if targetTag != "" {
		fullRef = fmt.Sprintf("%s:%s", fullRef, targetTag)
	}

	registryName := r.getRegistryName(ref.Registry)

	var manifest *ManifestInfo
	if metadata.Digest != "" {
		manifest = &ManifestInfo{
			Digest: metadata.Digest,
		}
	}

	return &ImageInfo{
		Name:         ref.Name,
		Description:  description,
		LatestTag:    targetTag,
		TagDate:      metadata.Date,
		Size:         metadata.Size,
		Manifest:     manifest,
		Registry:     registryName,
		FullImageRef: fullRef,
	}, nil
}

func (r *UnifiedRegistry) GetTags(ref ImageReference) ([]TagInfo, error) {
	repoRef, err := r.parseRef(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reference: %w", err)
	}

	tags, err := remote.List(repoRef, r.options...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	var result []TagInfo
	for _, tag := range tags {
		result = append(result, TagInfo{
			Name: tag,
		})
	}

	return sortTagsBySemver(result), nil
}

func (r *UnifiedRegistry) GetLatestTag(ref ImageReference) (string, error) {
	tags, err := r.GetTags(ref)
	if err != nil {
		return "", err
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found for image '%s'", ref.Name)
	}

	semverTags := filterSemverTags(tags)
	if len(semverTags) == 0 {
		return "", fmt.Errorf("no semantic version tags found for image '%s'", ref.Name)
	}

	return semverTags[0].Name, nil
}

func (r *UnifiedRegistry) GetBrowseURL(ref ImageReference) string {
	registry := ref.Registry
	if registry == "" || registry == RegistryDocker || registry == RegistryDockerV2 {
		org := ref.Organization
		if org == "" {
			org = "library"
		}
		return fmt.Sprintf("https://hub.docker.com/r/%s/%s", org, ref.Name)
	}

	switch {
	case strings.HasSuffix(registry, GHCR):
		return fmt.Sprintf("https://github.com/%s/%s/pkgs/container/%s", ref.Organization, ref.Name, ref.Name)
	case strings.HasSuffix(registry, GCR):
		project := ref.Organization
		if project == "" {
			project = registry
		}
		return fmt.Sprintf("https://console.cloud.google.com/gcr/images/%s/GLOBAL/%s", project, ref.Name)
	case strings.HasSuffix(registry, ECR):
		return fmt.Sprintf("https://gallery.ecr.aws/%s/%s", ref.Organization, ref.Name)
	case strings.HasSuffix(registry, ACR):
		if ref.Organization != "" {
			return fmt.Sprintf("https://%s/#/repository/%s/%s", registry, ref.Organization, ref.Name)
		}
		return fmt.Sprintf("https://%s/#/repository/%s", registry, ref.Name)
	case strings.HasSuffix(registry, Quay):
		return fmt.Sprintf("https://quay.io/repository/%s/%s", ref.Organization, ref.Name)
	default:
		return fmt.Sprintf("https://%s", registry)
	}
}

func (r *UnifiedRegistry) parseRef(ref ImageReference) (name.Repository, error) {
	imageStr := r.buildFullImageRef(ref)
	return name.NewRepository(imageStr)
}

func (r *UnifiedRegistry) fetchTagMetadata(repo name.Repository, tag string) (TagMetadata, error) {
	cacheKey := fmt.Sprintf("%s:%s", repo.Name(), tag)

	if desc, ok := r.cache[cacheKey]; ok {
		size, date := r.extractMetadataFromDesc(desc, repo, tag)
		return TagMetadata{
			Size:   size,
			Date:   date,
			Digest: desc.Digest.String(),
		}, nil
	}

	taggedRef, err := name.NewTag(cacheKey)
	if err != nil {
		return TagMetadata{}, fmt.Errorf("failed to create tag reference: %w", err)
	}

	desc, err := remote.Get(taggedRef, r.options...)
	if err != nil {
		return TagMetadata{}, fmt.Errorf("failed to fetch manifest from registry: %w", err)
	}

	r.cache[cacheKey] = desc

	size, date := r.extractMetadataFromDesc(desc, repo, tag)
	return TagMetadata{
		Size:   size,
		Date:   date,
		Digest: desc.Digest.String(),
	}, nil
}

func (r *UnifiedRegistry) isDockerHub(repo name.Repository) bool {
	registryStr := repo.Registry.Name()
	return registryStr == "" || registryStr == RegistryDocker || registryStr == RegistryDockerV2 || strings.Contains(registryStr, "docker.io")
}

func (r *UnifiedRegistry) extractMetadataFromDesc(desc *remote.Descriptor, repo name.Repository, tag string) (string, string) {
	var size string
	size, _ = r.calculateTotalImageSize(desc, repo)

	var date string
	if d, ok := desc.Annotations["org.opencontainers.image.created"]; ok && d != "" {
		date = d
	} else {
		img, err := desc.Image()
		if err == nil {
			configFile, err := img.ConfigFile()
			if err == nil && !configFile.Created.IsZero() {
				date = configFile.Created.Format(time.RFC3339)
			}
		}
	}

	if r.isDockerHub(repo) && tag != "" {
		if hubDate := r.fetchTagDateFromDockerHub(repo, tag); hubDate != "" {
			date = hubDate
		}
	}

	return size, date
}

func (r *UnifiedRegistry) calculateTotalImageSize(desc *remote.Descriptor, repo name.Repository) (string, error) {
	if desc.MediaType == types.DockerManifestList || desc.MediaType == types.OCIImageIndex {
		return r.calculateMultiArchSize(desc, repo)
	}

	img, err := desc.Image()
	if err != nil {
		return "", err
	}

	return r.calculateImageSize(img), nil
}

func (r *UnifiedRegistry) calculateImageSize(img v1.Image) string {
	layers, err := img.Layers()
	if err != nil {
		return ""
	}

	var totalSize int64
	for _, layer := range layers {
		size, err := layer.Size()
		if err == nil {
			totalSize += size
		}
	}

	configSize, _ := img.RawConfigFile()
	totalSize += int64(len(configSize))

	if totalSize > 0 {
		return formatBytes(totalSize)
	}

	return ""
}

func (r *UnifiedRegistry) calculateMultiArchSize(desc *remote.Descriptor, repo name.Repository) (string, error) {
	index, err := desc.ImageIndex()
	if err != nil {
		return "", err
	}

	manifest, err := index.IndexManifest()
	if err != nil {
		return "", err
	}

	var totalSize int64
	for _, platformDesc := range manifest.Manifests {
		if platformDesc.Platform != nil {
			digestRef := repo.Digest(platformDesc.Digest.String())
			img, err := remote.Image(digestRef, r.options...)
			if err != nil {
				continue
			}
			layers, err := img.Layers()
			if err != nil {
				continue
			}
			for _, layer := range layers {
				size, err := layer.Size()
				if err == nil {
					totalSize += size
				}
			}
			configSize, _ := img.RawConfigFile()
			totalSize += int64(len(configSize))
		}
	}

	if totalSize > 0 {
		return formatBytes(totalSize), nil
	}

	return "", nil
}

func (r *UnifiedRegistry) fetchDescription(ref ImageReference, imageName, targetTag string) string {
	if ref.Registry == "" || ref.Registry == RegistryDocker || ref.Registry == RegistryDockerV2 {
		if desc := r.fetchDockerHubDescription(ref); desc != "" {
			return desc
		}
	}

	if strings.HasSuffix(ref.Registry, Quay) {
		if desc := r.fetchQuayDescription(ref); desc != "" {
			return r.stripHTML(desc)
		}
	}

	if desc := r.fetchDescriptionFromLabels(imageName, targetTag); desc != "" {
		return desc
	}

	return "Description not available for this registry"
}

func (r *UnifiedRegistry) fetchDockerHubDescription(ref ImageReference) string {
	org := ref.Organization
	if org == "" {
		org = "library"
	}

	var result struct {
		Description string `json:"description"`
	}

	if r.fetchHTTP(fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/%s/", org, ref.Name), &result) {
		return result.Description
	}
	return ""
}

func (r *UnifiedRegistry) fetchQuayDescription(ref ImageReference) string {
	var result struct {
		Description string `json:"description"`
	}

	if r.fetchHTTP(fmt.Sprintf("https://quay.io/api/v1/repository/%s/%s", ref.Organization, ref.Name), &result) {
		return result.Description
	}
	return ""
}

func (r *UnifiedRegistry) fetchTagDateFromDockerHub(repo name.Repository, tag string) string {
	repoStr := repo.Name()

	if strings.Contains(repoStr, "/") {
		parts := strings.Split(repoStr, "/")
		if len(parts) >= 2 {
			repoStr = strings.Join(parts[1:], "/")
		}
	}

	parts := strings.Split(repoStr, "/")
	var org, name string
	if len(parts) == 1 {
		org = "library"
		name = parts[0]
	} else if len(parts) == 2 {
		org = parts[0]
		name = parts[1]
	} else {
		org = parts[0]
		name = strings.Join(parts[1:], "/")
	}

	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/%s/tags/?page_size=100", org, name)
	for url != "" {
		var pageResult struct {
			Next    string `json:"next"`
			Results []struct {
				Name        string `json:"name"`
				LastUpdated string `json:"last_updated"`
			} `json:"results"`
		}

		resp, err := r.httpClient.Get(url)
		if err != nil {
			return ""
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return ""
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return ""
		}

		if err := json.Unmarshal(body, &pageResult); err != nil {
			return ""
		}

		for _, t := range pageResult.Results {
			if t.Name == tag && t.LastUpdated != "" {
				return t.LastUpdated
			}
		}

		url = pageResult.Next
	}

	return ""
}

func (r *UnifiedRegistry) fetchHTTP(url string, target any) bool {
	resp, err := r.httpClient.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	if err := json.Unmarshal(body, target); err != nil {
		return false
	}

	return true
}

func (r *UnifiedRegistry) fetchDescriptionFromLabels(imageName, targetTag string) string {
	taggedRef, err := name.NewTag(fmt.Sprintf("%s:%s", imageName, targetTag))
	if err != nil {
		return ""
	}

	img, err := remote.Image(taggedRef, r.options...)
	if err != nil {
		return ""
	}

	configFile, err := img.ConfigFile()
	if err != nil {
		return ""
	}

	if len(configFile.Config.Labels) > 0 {
		if desc, ok := configFile.Config.Labels["org.opencontainers.image.description"]; ok && desc != "" {
			return desc
		}
		if desc, ok := configFile.Config.Labels["description"]; ok && desc != "" {
			return desc
		}
		if summary, ok := configFile.Config.Labels["org.opencontainers.image.title"]; ok && summary != "" {
			return summary
		}
	}

	return ""
}

func (r *UnifiedRegistry) stripHTML(html string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return strings.TrimSpace(re.ReplaceAllString(html, ""))
}

func (r *UnifiedRegistry) buildFullImageRef(ref ImageReference) string {
	if ref.Registry == "" {
		if ref.Organization == "" {
			return fmt.Sprintf("library/%s", ref.Name)
		}
		return fmt.Sprintf("%s/%s", ref.Organization, ref.Name)
	}

	if ref.Organization == "" {
		return fmt.Sprintf("%s/%s", ref.Registry, ref.Name)
	}

	return fmt.Sprintf("%s/%s/%s", ref.Registry, ref.Organization, ref.Name)
}

func (r *UnifiedRegistry) getRegistryName(registry string) string {
	if registry == "" || registry == RegistryDocker || registry == RegistryDockerV2 {
		return "Docker Hub"
	}

	switch {
	case strings.HasSuffix(registry, GHCR):
		return "GitHub Container Registry"
	case strings.HasSuffix(registry, GCR):
		return "Google Container Registry"
	case strings.HasSuffix(registry, ECR):
		return "Amazon Elastic Container Registry Public"
	case strings.HasSuffix(registry, ACR):
		return "Azure Container Registry"
	case strings.HasSuffix(registry, Quay):
		return "Quay.io"
	default:
		return registry
	}
}
