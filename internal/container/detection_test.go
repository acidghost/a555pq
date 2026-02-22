package container

import (
	"testing"
)

func TestDetectRegistry(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name             string
		image            string
		wantRegistry     string
		wantOrganization string
		wantName         string
		wantTag          string
	}{
		{
			name:             "Docker Hub - library image",
			image:            "nginx",
			wantRegistry:     "",
			wantOrganization: "",
			wantName:         "nginx",
			wantTag:          "",
		},
		{
			name:             "Docker Hub - library image with tag",
			image:            "nginx:latest",
			wantRegistry:     "",
			wantOrganization: "",
			wantName:         "nginx",
			wantTag:          "latest",
		},
		{
			name:             "Docker Hub - org image",
			image:            "library/nginx",
			wantRegistry:     "",
			wantOrganization: "library",
			wantName:         "nginx",
			wantTag:          "",
		},
		{
			name:             "Docker Hub - org image with tag",
			image:            "library/nginx:1.21",
			wantRegistry:     "",
			wantOrganization: "library",
			wantName:         "nginx",
			wantTag:          "1.21",
		},
		{
			name:             "Docker Hub - org slash image",
			image:            "nginxinc/nginx-unprivileged",
			wantRegistry:     "",
			wantOrganization: "nginxinc",
			wantName:         "nginx-unprivileged",
			wantTag:          "",
		},
		{
			name:             "GitHub Container Registry",
			image:            "ghcr.io/actions/checkout",
			wantRegistry:     "ghcr.io",
			wantOrganization: "actions",
			wantName:         "checkout",
			wantTag:          "",
		},
		{
			name:             "GitHub Container Registry with tag",
			image:            "ghcr.io/actions/checkout:v4",
			wantRegistry:     "ghcr.io",
			wantOrganization: "actions",
			wantName:         "checkout",
			wantTag:          "v4",
		},
		{
			name:             "Google Container Registry",
			image:            "gcr.io/google-containers/kube-proxy",
			wantRegistry:     "gcr.io",
			wantOrganization: "google-containers",
			wantName:         "kube-proxy",
			wantTag:          "",
		},
		{
			name:             "Google Container Registry with tag",
			image:            "gcr.io/distroless/static:nonroot",
			wantRegistry:     "gcr.io",
			wantOrganization: "distroless",
			wantName:         "static",
			wantTag:          "nonroot",
		},
		{
			name:             "Google Container Registry with project in path",
			image:            "gcr.io/kubernetes-helm/tiller",
			wantRegistry:     "gcr.io",
			wantOrganization: "kubernetes-helm",
			wantName:         "tiller",
			wantTag:          "",
		},
		{
			name:             "Amazon ECR Public",
			image:            "public.ecr.aws/nginx/nginx",
			wantRegistry:     "public.ecr.aws",
			wantOrganization: "nginx",
			wantName:         "nginx",
			wantTag:          "",
		},
		{
			name:             "Amazon ECR Public with tag",
			image:            "public.ecr.aws/nginx/nginx:latest",
			wantRegistry:     "public.ecr.aws",
			wantOrganization: "nginx",
			wantName:         "nginx",
			wantTag:          "latest",
		},
		{
			name:             "Azure Container Registry",
			image:            "myregistry.azurecr.io/myimage",
			wantRegistry:     "myregistry.azurecr.io",
			wantOrganization: "",
			wantName:         "myimage",
			wantTag:          "",
		},
		{
			name:             "Azure Container Registry with org and tag",
			image:            "myregistry.azurecr.io/myorg/myimage:1.0",
			wantRegistry:     "myregistry.azurecr.io",
			wantOrganization: "myorg",
			wantName:         "myimage",
			wantTag:          "1.0",
		},
		{
			name:             "Quay.io",
			image:            "quay.io/prometheus/prometheus",
			wantRegistry:     "quay.io",
			wantOrganization: "prometheus",
			wantName:         "prometheus",
			wantTag:          "",
		},
		{
			name:             "Quay.io with tag",
			image:            "quay.io/coreos/etcd:v3.5.0",
			wantRegistry:     "quay.io",
			wantOrganization: "coreos",
			wantName:         "etcd",
			wantTag:          "v3.5.0",
		},
		{
			name:             "Custom registry with IP",
			image:            "192.168.1.1:5000/myimage",
			wantRegistry:     "192.168.1.1:5000",
			wantOrganization: "",
			wantName:         "myimage",
			wantTag:          "",
		},
		{
			name:             "Custom registry with domain",
			image:            "registry.example.com/myorg/myimage",
			wantRegistry:     "registry.example.com",
			wantOrganization: "myorg",
			wantName:         "myimage",
			wantTag:          "",
		},
		{
			name:             "Custom registry with tag",
			image:            "registry.example.com/myorg/myimage:latest",
			wantRegistry:     "registry.example.com",
			wantOrganization: "myorg",
			wantName:         "myimage",
			wantTag:          "latest",
		},
		{
			name:             "Complex tag format",
			image:            "gcr.io/distroless/base:nonroot-debian11",
			wantRegistry:     "gcr.io",
			wantOrganization: "distroless",
			wantName:         "base",
			wantTag:          "nonroot-debian11",
		},
		{
			name:             "Nested organization path",
			image:            "ghcr.io/org/sub-org/repo",
			wantRegistry:     "ghcr.io",
			wantOrganization: "org",
			wantName:         "sub-org/repo",
			wantTag:          "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.detectRegistry(tt.image)
			if got.Registry != tt.wantRegistry {
				t.Errorf("detectRegistry() Registry = %v, want %v", got.Registry, tt.wantRegistry)
			}
			if got.Organization != tt.wantOrganization {
				t.Errorf("detectRegistry() Organization = %v, want %v", got.Organization, tt.wantOrganization)
			}
			if got.Name != tt.wantName {
				t.Errorf("detectRegistry() Name = %v, want %v", got.Name, tt.wantName)
			}
			if got.Tag != tt.wantTag {
				t.Errorf("detectRegistry() Tag = %v, want %v", got.Tag, tt.wantTag)
			}
		})
	}
}

func TestGetRegistryName(t *testing.T) {
	registry := NewUnifiedRegistry()

	tests := []struct {
		name     string
		registry string
		wantName string
	}{
		{
			name:     "Docker Hub - empty",
			registry: "",
			wantName: "Docker Hub",
		},
		{
			name:     "Docker Hub - docker.io",
			registry: "docker.io",
			wantName: "Docker Hub",
		},
		{
			name:     "GitHub Container Registry",
			registry: "ghcr.io",
			wantName: "GitHub Container Registry",
		},
		{
			name:     "GitHub Container Registry with subdomain",
			registry: "ghcr.io",
			wantName: "GitHub Container Registry",
		},
		{
			name:     "Google Container Registry",
			registry: "gcr.io",
			wantName: "Google Container Registry",
		},
		{
			name:     "Google Container Registry with project",
			registry: "us.gcr.io",
			wantName: "Google Container Registry",
		},
		{
			name:     "Amazon ECR Public",
			registry: "public.ecr.aws",
			wantName: "Amazon Elastic Container Registry Public",
		},
		{
			name:     "Amazon ECR Public - alternate",
			registry: "gallery.ecr.aws",
			wantName: "Amazon Elastic Container Registry Public",
		},
		{
			name:     "Azure Container Registry",
			registry: "myregistry.azurecr.io",
			wantName: "Azure Container Registry",
		},
		{
			name:     "Quay.io",
			registry: "quay.io",
			wantName: "Quay.io",
		},
		{
			name:     "Custom registry",
			registry: "registry.example.com",
			wantName: "registry.example.com",
		},
		{
			name:     "Custom registry with port",
			registry: "192.168.1.1:5000",
			wantName: "192.168.1.1:5000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.getRegistryName(tt.registry)
			if got != tt.wantName {
				t.Errorf("getRegistryName() = %v, want %v", got, tt.wantName)
			}
		})
	}
}

func TestGetBrowseURL(t *testing.T) {
	registry := NewUnifiedRegistry()

	tests := []struct {
		name    string
		ref     ImageReference
		wantURL string
	}{
		{
			name:    "Docker Hub - library image",
			ref:     ImageReference{Registry: "", Organization: "", Name: "nginx"},
			wantURL: "https://hub.docker.com/r/library/nginx",
		},
		{
			name:    "Docker Hub - org image",
			ref:     ImageReference{Registry: "", Organization: "nginxinc", Name: "nginx-unprivileged"},
			wantURL: "https://hub.docker.com/r/nginxinc/nginx-unprivileged",
		},
		{
			name:    "GitHub Container Registry",
			ref:     ImageReference{Registry: "ghcr.io", Organization: "actions", Name: "checkout"},
			wantURL: "https://github.com/actions/checkout/pkgs/container/checkout",
		},
		{
			name:    "Google Container Registry",
			ref:     ImageReference{Registry: "gcr.io", Organization: "google-containers", Name: "kube-proxy"},
			wantURL: "https://console.cloud.google.com/gcr/images/google-containers/GLOBAL/kube-proxy",
		},
		{
			name:    "Google Container Registry - no org",
			ref:     ImageReference{Registry: "us.gcr.io", Organization: "", Name: "myimage"},
			wantURL: "https://console.cloud.google.com/gcr/images/us.gcr.io/GLOBAL/myimage",
		},
		{
			name:    "Amazon ECR Public",
			ref:     ImageReference{Registry: "public.ecr.aws", Organization: "nginx", Name: "nginx"},
			wantURL: "https://gallery.ecr.aws/nginx/nginx",
		},
		{
			name:    "Azure Container Registry",
			ref:     ImageReference{Registry: "myregistry.azurecr.io", Organization: "", Name: "myimage"},
			wantURL: "https://myregistry.azurecr.io/#/repository/myimage",
		},
		{
			name:    "Azure Container Registry with org",
			ref:     ImageReference{Registry: "myregistry.azurecr.io", Organization: "myorg", Name: "myimage"},
			wantURL: "https://myregistry.azurecr.io/#/repository/myorg/myimage",
		},
		{
			name:    "Quay.io",
			ref:     ImageReference{Registry: "quay.io", Organization: "prometheus", Name: "prometheus"},
			wantURL: "https://quay.io/repository/prometheus/prometheus",
		},
		{
			name:    "Custom registry",
			ref:     ImageReference{Registry: "registry.example.com", Organization: "myorg", Name: "myimage"},
			wantURL: "https://registry.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.GetBrowseURL(tt.ref)
			if got != tt.wantURL {
				t.Errorf("GetBrowseURL() = %v, want %v", got, tt.wantURL)
			}
		})
	}
}

func TestBuildFullImageRef(t *testing.T) {
	registry := NewUnifiedRegistry()

	tests := []struct {
		name    string
		ref     ImageReference
		wantRef string
	}{
		{
			name:    "Docker Hub - library image",
			ref:     ImageReference{Registry: "", Organization: "", Name: "nginx"},
			wantRef: "library/nginx",
		},
		{
			name:    "Docker Hub - org image",
			ref:     ImageReference{Registry: "", Organization: "nginxinc", Name: "nginx-unprivileged"},
			wantRef: "nginxinc/nginx-unprivileged",
		},
		{
			name:    "GitHub Container Registry",
			ref:     ImageReference{Registry: "ghcr.io", Organization: "actions", Name: "checkout"},
			wantRef: "ghcr.io/actions/checkout",
		},
		{
			name:    "Google Container Registry",
			ref:     ImageReference{Registry: "gcr.io", Organization: "google-containers", Name: "kube-proxy"},
			wantRef: "gcr.io/google-containers/kube-proxy",
		},
		{
			name:    "Amazon ECR Public",
			ref:     ImageReference{Registry: "public.ecr.aws", Organization: "nginx", Name: "nginx"},
			wantRef: "public.ecr.aws/nginx/nginx",
		},
		{
			name:    "Azure Container Registry - no org",
			ref:     ImageReference{Registry: "myregistry.azurecr.io", Organization: "", Name: "myimage"},
			wantRef: "myregistry.azurecr.io/myimage",
		},
		{
			name:    "Quay.io",
			ref:     ImageReference{Registry: "quay.io", Organization: "prometheus", Name: "prometheus"},
			wantRef: "quay.io/prometheus/prometheus",
		},
		{
			name:    "Custom registry with org",
			ref:     ImageReference{Registry: "registry.example.com", Organization: "myorg", Name: "myimage"},
			wantRef: "registry.example.com/myorg/myimage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.buildFullImageRef(tt.ref)
			if got != tt.wantRef {
				t.Errorf("buildFullImageRef() = %v, want %v", got, tt.wantRef)
			}
		})
	}
}
