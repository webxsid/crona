package updatecheck

import (
	"strconv"
	"strings"
)

type semver struct {
	major      int
	minor      int
	patch      int
	prerelease string
}

func isNewerVersion(current, latest string) bool {
	currentVersion, ok := parseSemver(current)
	if !ok {
		return false
	}
	latestVersion, ok := parseSemver(latest)
	if !ok {
		return false
	}
	if currentVersion.major != latestVersion.major {
		return latestVersion.major > currentVersion.major
	}
	if currentVersion.minor != latestVersion.minor {
		return latestVersion.minor > currentVersion.minor
	}
	if currentVersion.patch != latestVersion.patch {
		return latestVersion.patch > currentVersion.patch
	}
	if currentVersion.prerelease == latestVersion.prerelease {
		return false
	}
	if currentVersion.prerelease == "" && latestVersion.prerelease != "" {
		return false
	}
	if currentVersion.prerelease != "" && latestVersion.prerelease == "" {
		return true
	}
	return comparePrerelease(latestVersion.prerelease, currentVersion.prerelease) > 0
}

func compareSemver(left, right semver) int {
	if left.major != right.major {
		if left.major > right.major {
			return 1
		}
		return -1
	}
	if left.minor != right.minor {
		if left.minor > right.minor {
			return 1
		}
		return -1
	}
	if left.patch != right.patch {
		if left.patch > right.patch {
			return 1
		}
		return -1
	}
	if left.prerelease == right.prerelease {
		return 0
	}
	if left.prerelease == "" {
		return 1
	}
	if right.prerelease == "" {
		return -1
	}
	return comparePrerelease(left.prerelease, right.prerelease)
}

func normalizeVersion(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "v")
	if idx := strings.IndexByte(value, '+'); idx >= 0 {
		value = value[:idx]
	}
	return value
}

func parseSemver(value string) (semver, bool) {
	value = normalizeVersion(value)
	if value == "" {
		return semver{}, false
	}
	var prerelease string
	if idx := strings.IndexByte(value, '-'); idx >= 0 {
		prerelease = value[idx+1:]
		value = value[:idx]
	}
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return semver{}, false
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return semver{}, false
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return semver{}, false
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return semver{}, false
	}
	return semver{major: major, minor: minor, patch: patch, prerelease: prerelease}, true
}

func comparePrerelease(left, right string) int {
	leftParts := strings.Split(left, ".")
	rightParts := strings.Split(right, ".")
	limit := len(leftParts)
	if len(rightParts) < limit {
		limit = len(rightParts)
	}
	for i := 0; i < limit; i++ {
		if cmp := comparePrereleasePart(leftParts[i], rightParts[i]); cmp != 0 {
			return cmp
		}
	}
	switch {
	case len(leftParts) > len(rightParts):
		return 1
	case len(leftParts) < len(rightParts):
		return -1
	default:
		return 0
	}
}

func comparePrereleasePart(left, right string) int {
	leftNum, leftErr := strconv.Atoi(left)
	rightNum, rightErr := strconv.Atoi(right)
	switch {
	case leftErr == nil && rightErr == nil:
		switch {
		case leftNum > rightNum:
			return 1
		case leftNum < rightNum:
			return -1
		default:
			return 0
		}
	case leftErr == nil:
		return -1
	case rightErr == nil:
		return 1
	case left > right:
		return 1
	case left < right:
		return -1
	default:
		return 0
	}
}
