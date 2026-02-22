package container

import (
	"sort"
	"strings"

	"github.com/Masterminds/semver"
)

func parseSemver(tag string) *semver.Version {
	cleanTag := strings.TrimPrefix(tag, "v")
	cleanTag = strings.TrimSuffix(cleanTag, "-alpine")
	cleanTag = strings.TrimSuffix(cleanTag, "-slim")
	cleanTag = strings.TrimSuffix(cleanTag, "-buster")
	cleanTag = strings.TrimSuffix(cleanTag, "-bullseye")
	cleanTag = strings.TrimSuffix(cleanTag, "-bookworm")
	cleanTag = strings.TrimSuffix(cleanTag, "-focal")
	cleanTag = strings.TrimSuffix(cleanTag, "-jammy")

	version, err := semver.NewVersion(cleanTag)
	if err != nil {
		return nil
	}
	return version
}

func filterSemverTags(tags []TagInfo) []TagInfo {
	var semverTags []TagInfo
	for _, tag := range tags {
		if parseSemver(tag.Name) != nil {
			semverTags = append(semverTags, tag)
		}
	}
	return sortTagsBySemver(semverTags)
}

func sortTagsBySemver(tags []TagInfo) []TagInfo {
	result := make([]TagInfo, len(tags))
	copy(result, tags)

	sort.Slice(result, func(i, j int) bool {
		va := parseSemver(result[i].Name)
		vb := parseSemver(result[j].Name)

		if va != nil && vb != nil {
			if va.Equal(vb) {
				return result[i].Name < result[j].Name
			}
			return vb.LessThan(va)
		}

		if va != nil {
			return false
		}

		if vb != nil {
			return true
		}

		return result[i].Name > result[j].Name
	})

	return result
}
