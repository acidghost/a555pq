package container

import (
	"net/http"
	"testing"
	"time"
)

func TestGetBrowseURL_DockerHub(t *testing.T) {
	registry := NewUnifiedRegistry()

	tests := []struct {
		name    string
		ref     ImageReference
		wantURL string
	}{
		{
			name:    "library image",
			ref:     ImageReference{Registry: "", Organization: "", Name: "nginx"},
			wantURL: "https://hub.docker.com/r/library/nginx",
		},
		{
			name:    "library image with docker.io registry",
			ref:     ImageReference{Registry: "docker.io", Organization: "", Name: "nginx"},
			wantURL: "https://hub.docker.com/r/library/nginx",
		},
		{
			name:    "library image with registry-1.docker.io registry",
			ref:     ImageReference{Registry: "registry-1.docker.io", Organization: "", Name: "nginx"},
			wantURL: "https://hub.docker.com/r/library/nginx",
		},
		{
			name:    "org image",
			ref:     ImageReference{Registry: "", Organization: "nginxinc", Name: "nginx-unprivileged"},
			wantURL: "https://hub.docker.com/r/nginxinc/nginx-unprivileged",
		},
		{
			name:    "org image with explicit library org",
			ref:     ImageReference{Registry: "", Organization: "library", Name: "nginx"},
			wantURL: "https://hub.docker.com/r/library/nginx",
		},
		{
			name:    "complex org name",
			ref:     ImageReference{Registry: "", Organization: "my-org-name", Name: "my-image"},
			wantURL: "https://hub.docker.com/r/my-org-name/my-image",
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

func TestGetBrowseURL_GHCR(t *testing.T) {
	registry := NewUnifiedRegistry()

	tests := []struct {
		name    string
		ref     ImageReference
		wantURL string
	}{
		{
			name:    "simple org and image",
			ref:     ImageReference{Registry: "ghcr.io", Organization: "acidghost", Name: "a555cribe"},
			wantURL: "https://github.com/acidghost/a555cribe/pkgs/container/a555cribe",
		},
		{
			name:    "coder code-server",
			ref:     ImageReference{Registry: "ghcr.io", Organization: "coder", Name: "code-server"},
			wantURL: "https://github.com/coder/code-server/pkgs/container/code-server",
		},
		{
			name:    "actions checkout",
			ref:     ImageReference{Registry: "ghcr.io", Organization: "actions", Name: "checkout"},
			wantURL: "https://github.com/actions/checkout/pkgs/container/checkout",
		},
		{
			name:    "github actions python-versions",
			ref:     ImageReference{Registry: "ghcr.io", Organization: "actions", Name: "python-versions"},
			wantURL: "https://github.com/actions/python-versions/pkgs/container/python-versions",
		},
		{
			name:    "org with hyphen",
			ref:     ImageReference{Registry: "ghcr.io", Organization: "my-org", Name: "my-image"},
			wantURL: "https://github.com/my-org/my-image/pkgs/container/my-image",
		},
		{
			name:    "complex image name",
			ref:     ImageReference{Registry: "ghcr.io", Organization: "my-org", Name: "my-complex-image-name"},
			wantURL: "https://github.com/my-org/my-complex-image-name/pkgs/container/my-complex-image-name",
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

func TestGetBrowseURL_GCR(t *testing.T) {
	registry := NewUnifiedRegistry()

	tests := []struct {
		name    string
		ref     ImageReference
		wantURL string
	}{
		{
			name:    "distroless static",
			ref:     ImageReference{Registry: "gcr.io", Organization: "distroless", Name: "static"},
			wantURL: "https://console.cloud.google.com/gcr/images/distroless/GLOBAL/static",
		},
		{
			name:    "google-containers kube-proxy",
			ref:     ImageReference{Registry: "gcr.io", Organization: "google-containers", Name: "kube-proxy"},
			wantURL: "https://console.cloud.google.com/gcr/images/google-containers/GLOBAL/kube-proxy",
		},
		{
			name:    "no organization - uses registry",
			ref:     ImageReference{Registry: "us.gcr.io", Organization: "", Name: "myimage"},
			wantURL: "https://console.cloud.google.com/gcr/images/us.gcr.io/GLOBAL/myimage",
		},
		{
			name:    "eu.gcr.io with org",
			ref:     ImageReference{Registry: "eu.gcr.io", Organization: "my-project", Name: "my-image"},
			wantURL: "https://console.cloud.google.com/gcr/images/my-project/GLOBAL/my-image",
		},
		{
			name:    "asia.gcr.io",
			ref:     ImageReference{Registry: "asia.gcr.io", Organization: "my-project", Name: "app"},
			wantURL: "https://console.cloud.google.com/gcr/images/my-project/GLOBAL/app",
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

func TestGetBrowseURL_Quay(t *testing.T) {
	registry := NewUnifiedRegistry()

	tests := []struct {
		name    string
		ref     ImageReference
		wantURL string
	}{
		{
			name:    "prometheus prometheus",
			ref:     ImageReference{Registry: "quay.io", Organization: "prometheus", Name: "prometheus"},
			wantURL: "https://quay.io/repository/prometheus/prometheus",
		},
		{
			name:    "coreos etcd",
			ref:     ImageReference{Registry: "quay.io", Organization: "coreos", Name: "etcd"},
			wantURL: "https://quay.io/repository/coreos/etcd",
		},
		{
			name:    "with org",
			ref:     ImageReference{Registry: "quay.io", Organization: "myorg", Name: "myimage"},
			wantURL: "https://quay.io/repository/myorg/myimage",
		},
		{
			name:    "complex org name",
			ref:     ImageReference{Registry: "quay.io", Organization: "my-org-team", Name: "my-app"},
			wantURL: "https://quay.io/repository/my-org-team/my-app",
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

func TestGetBrowseURL_ECR(t *testing.T) {
	registry := NewUnifiedRegistry()

	tests := []struct {
		name    string
		ref     ImageReference
		wantURL string
	}{
		{
			name:    "nginx nginx",
			ref:     ImageReference{Registry: "public.ecr.aws", Organization: "nginx", Name: "nginx"},
			wantURL: "https://gallery.ecr.aws/nginx/nginx",
		},
		{
			name:    "amazon aws-nodejs",
			ref:     ImageReference{Registry: "public.ecr.aws", Organization: "amazon", Name: "aws-nodejs"},
			wantURL: "https://gallery.ecr.aws/amazon/aws-nodejs",
		},
		{
			name:    "amazonlinux amazonlinux",
			ref:     ImageReference{Registry: "public.ecr.aws", Organization: "amazonlinux", Name: "amazonlinux"},
			wantURL: "https://gallery.ecr.aws/amazonlinux/amazonlinux",
		},
		{
			name:    "emergingtech emulated-arm64",
			ref:     ImageReference{Registry: "public.ecr.aws", Organization: "emergingtech", Name: "emulated-arm64"},
			wantURL: "https://gallery.ecr.aws/emergingtech/emulated-arm64",
		},
		{
			name:    "cplnbackend cpln-base-image",
			ref:     ImageReference{Registry: "public.ecr.aws", Organization: "cplnbackend", Name: "cpln-base-image"},
			wantURL: "https://gallery.ecr.aws/cplnbackend/cpln-base-image",
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

func TestGetBrowseURL_ACR(t *testing.T) {
	registry := NewUnifiedRegistry()

	tests := []struct {
		name    string
		ref     ImageReference
		wantURL string
	}{
		{
			name:    "no organization",
			ref:     ImageReference{Registry: "myregistry.azurecr.io", Organization: "", Name: "myimage"},
			wantURL: "https://myregistry.azurecr.io/#/repository/myimage",
		},
		{
			name:    "with organization",
			ref:     ImageReference{Registry: "myregistry.azurecr.io", Organization: "myorg", Name: "myimage"},
			wantURL: "https://myregistry.azurecr.io/#/repository/myorg/myimage",
		},
		{
			name:    "sub organization",
			ref:     ImageReference{Registry: "myregistry.azurecr.io", Organization: "myorg/myteam", Name: "myimage"},
			wantURL: "https://myregistry.azurecr.io/#/repository/myorg/myteam/myimage",
		},
		{
			name:    "complex image name",
			ref:     ImageReference{Registry: "myregistry.azurecr.io", Organization: "myorg", Name: "my-complex-image"},
			wantURL: "https://myregistry.azurecr.io/#/repository/myorg/my-complex-image",
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

func TestGetBrowseURL_CustomRegistry(t *testing.T) {
	registry := NewUnifiedRegistry()

	tests := []struct {
		name    string
		ref     ImageReference
		wantURL string
	}{
		{
			name:    "custom domain",
			ref:     ImageReference{Registry: "registry.example.com", Organization: "myorg", Name: "myimage"},
			wantURL: "https://registry.example.com",
		},
		{
			name:    "custom registry with port",
			ref:     ImageReference{Registry: "192.168.1.1:5000", Organization: "", Name: "myimage"},
			wantURL: "https://192.168.1.1:5000",
		},
		{
			name:    "localhost registry",
			ref:     ImageReference{Registry: "localhost:5000", Organization: "", Name: "myimage"},
			wantURL: "https://localhost:5000",
		},
		{
			name:    "custom organization registry",
			ref:     ImageReference{Registry: "registry.mycompany.com", Organization: "team", Name: "app"},
			wantURL: "https://registry.mycompany.com",
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

func TestGetBrowseURL_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	tests := []struct {
		name          string
		image         string
		expectSuccess bool
		skip          bool
		skipReason    string
	}{
		{
			name:          "Docker Hub - nginx",
			image:         "nginx",
			expectSuccess: true,
		},
		{
			name:          "Docker Hub - nginxinc/nginx-unprivileged",
			image:         "nginxinc/nginx-unprivileged",
			expectSuccess: true,
		},
		{
			name:          "GCR - distroless/static",
			image:         "gcr.io/distroless/static",
			expectSuccess: true,
		},
		{
			name:          "Quay - prometheus/prometheus",
			image:         "quay.io/prometheus/prometheus",
			expectSuccess: true,
		},
		{
			name:          "ECR - nginx/nginx",
			image:         "public.ecr.aws/nginx/nginx",
			expectSuccess: true,
		},
		{
			name:          "GHCR - coder/code-server",
			image:         "ghcr.io/coder/code-server",
			expectSuccess: true,
		},
		{
			name:          "GHCR - acidghost/a555cribe",
			image:         "ghcr.io/acidghost/a555cribe",
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip(tt.skipReason)
			}

			containerClient := NewClient()
			url, err := containerClient.GetBrowseURL(tt.image)
			if err != nil {
				t.Fatalf("GetBrowseURL() error = %v", err)
			}

			resp, err := client.Get(url)
			if err != nil {
				t.Errorf("Failed to fetch URL %s: %v", url, err)
				return
			}
			defer resp.Body.Close()

			success := resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusFound

			if tt.expectSuccess && !success {
				t.Errorf("Expected success for image %s, got status %d for URL %s", tt.image, resp.StatusCode, url)
			}
		})
	}
}

func TestGetBrowseURL_AllRegistries(t *testing.T) {
	registry := NewUnifiedRegistry()

	allTests := []struct {
		name string
		ref  ImageReference
	}{
		{
			name: "Docker Hub library",
			ref:  ImageReference{Registry: "", Organization: "", Name: "nginx"},
		},
		{
			name: "Docker Hub org",
			ref:  ImageReference{Registry: "", Organization: "nginxinc", Name: "nginx-unprivileged"},
		},
		{
			name: "GHCR",
			ref:  ImageReference{Registry: "ghcr.io", Organization: "acidghost", Name: "a555cribe"},
		},
		{
			name: "GCR",
			ref:  ImageReference{Registry: "gcr.io", Organization: "distroless", Name: "static"},
		},
		{
			name: "Quay",
			ref:  ImageReference{Registry: "quay.io", Organization: "prometheus", Name: "prometheus"},
		},
		{
			name: "ECR",
			ref:  ImageReference{Registry: "public.ecr.aws", Organization: "nginx", Name: "nginx"},
		},
		{
			name: "ACR",
			ref:  ImageReference{Registry: "myregistry.azurecr.io", Organization: "myorg", Name: "myimage"},
		},
		{
			name: "Custom",
			ref:  ImageReference{Registry: "registry.example.com", Organization: "myorg", Name: "myimage"},
		},
	}

	for _, tt := range allTests {
		t.Run(tt.name, func(t *testing.T) {
			url := registry.GetBrowseURL(tt.ref)
			if url == "" {
				t.Error("GetBrowseURL() returned empty string")
			}
			if url[0:8] != "https://" {
				t.Errorf("GetBrowseURL() URL doesn't start with https://: %s", url)
			}
		})
	}
}
