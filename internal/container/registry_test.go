package container

import "testing"

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
