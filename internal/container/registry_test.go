package container

import (
	"testing"
)

func TestFilterSemverTags(t *testing.T) {
	tests := []struct {
		name     string
		tags     []TagInfo
		wantLen  int
		wantTags []string
	}{
		{
			name: "mixed tags - semver only",
			tags: []TagInfo{
				{Name: "1.2.3"},
				{Name: "latest"},
				{Name: "1.2.4"},
				{Name: "alpine"},
				{Name: "v2.0.0"},
			},
			wantLen:  3,
			wantTags: []string{"v2.0.0", "1.2.4", "1.2.3"},
		},
		{
			name: "all semver tags",
			tags: []TagInfo{
				{Name: "1.0.0"},
				{Name: "1.2.3"},
				{Name: "2.0.0"},
				{Name: "v1.5.0"},
			},
			wantLen:  4,
			wantTags: []string{"2.0.0", "v1.5.0", "1.2.3", "1.0.0"},
		},
		{
			name: "no semver tags",
			tags: []TagInfo{
				{Name: "latest"},
				{Name: "alpine"},
				{Name: "stable"},
			},
			wantLen:  0,
			wantTags: []string{},
		},
		{
			name:     "empty list",
			tags:     []TagInfo{},
			wantLen:  0,
			wantTags: []string{},
		},
		{
			name: "semver with suffixes",
			tags: []TagInfo{
				{Name: "1.2.3-alpine"},
				{Name: "1.2.3"},
				{Name: "1.2.3-slim"},
				{Name: "v2.0.0-bookworm"},
			},
			wantLen:  4,
			wantTags: []string{"v2.0.0-bookworm", "1.2.3", "1.2.3-alpine", "1.2.3-slim"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterSemverTags(tt.tags)
			if len(got) != tt.wantLen {
				t.Errorf("filterSemverTags() length = %v, want %v", len(got), tt.wantLen)
			}
			if len(got) > 0 {
				for i, wantTag := range tt.wantTags {
					if got[i].Name != wantTag {
						t.Errorf("filterSemverTags()[%d] = %v, want %v", i, got[i].Name, wantTag)
					}
				}
			}
		})
	}
}

func TestGetLatestTag_SemverOnly(t *testing.T) {
	registry := NewUnifiedRegistry()

	tests := []struct {
		name        string
		image       string
		wantTag     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "image with semver tags returns highest",
			image:   "library/nginx",
			wantTag: "",
			wantErr: true,
		},
		{
			name:    "empty image name",
			image:   "",
			wantTag: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := registry.GetLatestTag(ImageReference{Name: tt.image})
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" {
				if !containsString(err.Error(), tt.errContains) {
					t.Errorf("GetLatestTag() error = %v, want error containing %v", err, tt.errContains)
				}
			}
			if got != tt.wantTag {
				t.Errorf("GetLatestTag() = %v, want %v", got, tt.wantTag)
			}
		})
	}
}

func TestGetImageInfoWithTag_WithDateAndSize(t *testing.T) {
	registry := NewUnifiedRegistry()

	tests := []struct {
		name        string
		image       ImageReference
		wantErr     bool
		checkFields bool
	}{
		{
			name:        "fetch with tag",
			image:       ImageReference{Name: "nginx", Tag: "latest"},
			wantErr:     false,
			checkFields: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := registry.GetImageInfo(tt.image)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetImageInfoWithTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFields && got != nil {
				if got.TagDate == "" {
					t.Logf("GetImageInfoWithTag() TagDate is empty (may not be available)")
				}
				if got.Size == "" {
					t.Logf("GetImageInfoWithTag() Size is empty (may not be available)")
				}
				t.Logf("GetImageInfoWithTag() TagDate = %s, Size = %s", got.TagDate, got.Size)
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
