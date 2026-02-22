package container

import (
	"strings"
)

const (
	DockerHub        = ""
	RegistryDocker   = "docker.io"
	RegistryDockerV2 = "registry-1.docker.io"
	GHCR             = "ghcr.io"
	GCR              = "gcr.io"
	ECR              = "ecr.aws"
	ACR              = "azurecr.io"
	Quay             = "quay.io"
)

type Client struct {
	registry *UnifiedRegistry
}

func NewClient() *Client {
	return &Client{
		registry: NewUnifiedRegistry(),
	}
}

func (c *Client) detectRegistry(image string) ImageReference {
	var ref ImageReference

	lastColon := strings.LastIndex(image, ":")
	lastSlash := strings.LastIndex(image, "/")

	pathPart := image
	if lastColon > lastSlash {
		ref.Tag = image[lastColon+1:]
		pathPart = image[:lastColon]
	}

	segments := strings.Split(pathPart, "/")

	if len(segments) == 1 {
		ref.Name = segments[0]
		return ref
	}

	first := segments[0]
	rest := strings.Join(segments[1:], "/")

	if strings.Contains(first, ".") || strings.Contains(first, ":") {
		ref.Registry = first
		c.parseRestParts(rest, &ref)
		return ref
	}

	for _, prefix := range []string{GHCR, GCR, ECR, ACR, Quay} {
		if strings.HasPrefix(first, prefix) {
			if strings.Contains(first, "/") {
				prefixParts := strings.SplitN(first, "/", 2)
				ref.Registry = prefixParts[0]
				ref.Organization = prefixParts[1]
			} else {
				ref.Registry = first
			}
			c.parseRestParts(rest, &ref)
			return ref
		}
	}

	ref.Organization = segments[0]
	restParts := strings.SplitN(rest, "/", 2)
	if len(restParts) == 2 {
		ref.Name = restParts[1]
	} else {
		ref.Name = restParts[0]
	}

	return ref
}

func (c *Client) parseRestParts(rest string, ref *ImageReference) {
	restParts := strings.SplitN(rest, "/", 2)
	if len(restParts) == 2 {
		if ref.Organization == "" {
			ref.Organization = restParts[0]
		}
		ref.Name = restParts[1]
	} else {
		if ref.Name == "" {
			ref.Name = restParts[0]
		}
	}
}

func (c *Client) GetImageInfo(image string) (*ImageInfo, error) {
	ref := c.detectRegistry(image)
	return c.registry.GetImageInfo(ref)
}

func (c *Client) GetTags(image string) ([]TagInfo, error) {
	ref := c.detectRegistry(image)
	return c.registry.GetTags(ref)
}

func (c *Client) GetLatestTag(image string) (string, error) {
	ref := c.detectRegistry(image)
	return c.registry.GetLatestTag(ref)
}

func (c *Client) GetBrowseURL(image string) (string, error) {
	ref := c.detectRegistry(image)
	return c.registry.GetBrowseURL(ref), nil
}
