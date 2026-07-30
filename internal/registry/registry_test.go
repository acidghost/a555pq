package registry

import (
	"testing"
	"time"

	"github.com/git-pkgs/registries"
)

func v(number string, publishedAt time.Time, status registries.VersionStatus) registries.Version {
	return registries.Version{Number: number, PublishedAt: publishedAt, Status: status}
}

func TestFilterByMinReleaseAgeAt(t *testing.T) {
	now := time.Date(2025, 1, 31, 12, 0, 0, 0, time.UTC)
	day := 24 * time.Hour

	tests := []struct {
		name     string
		versions []registries.Version
		minAge   time.Duration
		want     []string
	}{
		{
			name:     "all older than window are kept",
			minAge:   7 * day,
			versions: []registries.Version{v("1.0.0", now.Add(-30*day), ""), v("1.1.0", now.Add(-8*day), "")},
			want:     []string{"1.0.0", "1.1.0"},
		},
		{
			name:     "recent versions filtered out",
			minAge:   7 * day,
			versions: []registries.Version{v("1.0.0", now.Add(-30*day), ""), v("1.1.0", now.Add(-1*day), "")},
			want:     []string{"1.0.0"},
		},
		{
			name:     "version released exactly at the cutoff is kept",
			minAge:   7 * day,
			versions: []registries.Version{v("1.0.0", now.Add(-7*day), "")},
			want:     []string{"1.0.0"},
		},
		{
			name:     "versions without timestamp are kept",
			minAge:   7 * day,
			versions: []registries.Version{v("1.0.0", time.Time{}, ""), v("1.1.0", now.Add(-1*day), "")},
			want:     []string{"1.0.0"},
		},
		{
			name:     "disabled when min age is zero",
			minAge:   0,
			versions: []registries.Version{v("1.0.0", now.Add(-1*day), ""), v("1.1.0", now, "")},
			want:     []string{"1.0.0", "1.1.0"},
		},
		{
			name:     "disabled when min age is negative",
			minAge:   -5 * day,
			versions: []registries.Version{v("1.0.0", now.Add(-1*day), "")},
			want:     []string{"1.0.0"},
		},
		{
			name:     "empty input",
			minAge:   7 * day,
			versions: nil,
			want:     nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterByMinReleaseAgeAt(tt.versions, tt.minAge, now)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d versions %v, want %d %v", len(got), numbers(got), len(tt.want), tt.want)
			}
			for i, ver := range got {
				if ver.Number != tt.want[i] {
					t.Errorf("got[%d] = %v, want %s", i, ver.Number, tt.want[i])
				}
			}
		})
	}
}

func TestSelectLatest(t *testing.T) {
	day := 24 * time.Hour
	now := time.Now().UTC()

	tests := []struct {
		name       string
		versions   []registries.Version
		wantNumber string
		wantNil    bool
	}{
		{
			name:       "newest by publish time",
			versions:   []registries.Version{v("1.0.0", now.Add(-30*day), ""), v("1.1.0", now.Add(-1*day), "")},
			wantNumber: "1.1.0",
		},
		{
			name:       "skips yanked versions",
			versions:   []registries.Version{v("1.1.0", now.Add(-1*day), registries.StatusYanked), v("1.0.0", now.Add(-30*day), "")},
			wantNumber: "1.0.0",
		},
		{
			name:       "skips deprecated and retracted",
			versions:   []registries.Version{v("2.0.0", now.Add(-1*day), registries.StatusDeprecated), v("1.5.0", now.Add(-2*day), registries.StatusRetracted), v("1.0.0", now.Add(-30*day), "")},
			wantNumber: "1.0.0",
		},
		{
			name:       "falls back to list order when no timestamps",
			versions:   []registries.Version{v("3.0.0", time.Time{}, ""), v("2.0.0", time.Time{}, "")},
			wantNumber: "3.0.0",
		},
		{
			name:     "all yanked returns nil",
			versions: []registries.Version{v("1.0.0", now.Add(-1*day), registries.StatusYanked)},
			wantNil:  true,
		},
		{
			name:    "empty returns nil",
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectLatest(tt.versions)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("got %v, want nil", got.Number)
				}
				return
			}
			if got == nil {
				t.Fatalf("got nil, want %s", tt.wantNumber)
			}
			if got.Number != tt.wantNumber {
				t.Errorf("got %s, want %s", got.Number, tt.wantNumber)
			}
		})
	}
}

func numbers(vs []registries.Version) []string {
	out := make([]string, len(vs))
	for i, v := range vs {
		out[i] = v.Number
	}
	return out
}
