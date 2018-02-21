package phpcompat

import (
	"strings"
	"strconv"
	"fmt"
	"github.com/blang/semver"
	"sort"
)

var (
	phpVersions = map[string]map[string]string{
		"5.2": {
			"max": "5.2.17",
		},
		"5.3": {
			"max": "5.3.29",
		},
		"5.4": {
			"max": "5.4.45",
		},
		"5.5": {
			"max": "5.5.37",
		},
		"5.6": {
			"max": "5.6.32",
		},
		"7.0": {
			"max": "7.0.26",
		},
		"7.1": {
			"max": "7.1.13",
		},
		"7.2": {
			"max": "7.2.1",
		},
	}
	PhpLatest = "7.2.1"
)

type Compatibility struct {
	Source string              `json:"source"`
	Breaks *CompatibilityRange `json:"breaks,omitempty"`
	Warns  *CompatibilityRange `json:"warns,omitempty"`
}

type CompatibilityRange struct {
	Low        string `json:"low"`
	High       string `json:"high"`
	Reported   string `json:"reported"`
	MajorMinor string `json:"major_minor"`
}

// PreviousVersion returns the immediate previous version given a version.
func PreviousVersion(version string) string {

	if version == "all" {
		return version
	}

	// Only supporting down to 5.2.0.
	if version == "5.2" || version == "5.2.0" {
		return "5.2.0"
	}

	parts := strings.Split(version, ".")
	if len(parts) < 3 {
		parts = append(parts, "0")
	}

	var maxPrev []string

	if parts[1] == "0" {
		var mMPre string
		switch parts[0] + "." + parts[1] {
		case "7.0":
			mMPre = "5.6"
		}
		maxPrev = strings.Split(phpVersions[mMPre]["max"], ".")
	} else {
		pre, _ := strconv.Atoi(parts[1])
		pre -= 1
		maxPrev = strings.Split(phpVersions[parts[0]+"."+strconv.Itoa(pre)]["max"], ".")
	}

	// Convert and subtract parts
	p3, _ := strconv.Atoi(parts[2])
	p3 -= 1
	if p3 < 0 {
		parts[2] = maxPrev[2]

		if parts[1] == "0" {
			parts[1] = maxPrev[1]
			parts[0] = maxPrev[0]
		} else {

			p2, _ := strconv.Atoi(parts[1])
			p2 -= 1
			parts[1] = strconv.Itoa(p2)

		}
	} else {
		parts[2] = strconv.Itoa(p3)
	}

	return strings.Join(parts, ".")
}

// VersionParts takes a version given as string and returns each part as an int.
func VersionParts(version string) (int, int, int) {
	if version == "all" {
		return 0, 0, 0
	}
	parts := strings.Split(version, ".")

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	return major, minor, patch
}

// GerVersionParts returns a version range (low and high) as well as the majorMinor for the given version and the given version as `reported`.
func GetVersionParts(version, lowIn string) (low, high, majorMinor, reported string) {
	vParts := strings.Split(version, ".")

	majorMinor = "all"
	high = ""

	// Version "all" is not helpful.
	if len(vParts) > 1 {
		majorMinor = fmt.Sprintf("%s.%s", vParts[0], vParts[1])
	}

	// Is it a Major.Minor? Then get the max.
	if len(vParts) != 1 && len(vParts) != 3 {
		high = phpVersions[majorMinor]["max"]
	} else {
		high = version
	}

	if lowIn == "" {
		if majorMinor != "all" {
			low = fmt.Sprintf("%s.%s", majorMinor, "0")
		} else {
			low = "all"
		}

	} else {
		lowParts := strings.Split(lowIn, ".")
		if len(lowParts) == 2 {
			low = fmt.Sprintf("%s.%s", lowIn, "0")
		} else {
			low = lowIn
		}
	}

	v, _ := semver.Make(majorMinor + ".0")
	min, _ := semver.Make("5.2.0")

	if v.LT(min) && majorMinor != "all" {
		low = "5.2.0"
		high = low
		reported = majorMinor
		majorMinor = "5.2"
	}

	reported = version

	return
}

// BreasksVersions takes a PHPCompatibility sniff code and returns the versions that break for that code.
func BreaksVersions(code string) []string {

	compat, ok := PhpCompatibilityMap[code]

	if ! ok {
		return nil
	}

	broken := []string{}

	if compat.Breaks == nil {
		return nil
	}

	var rangeString string
	if compat.Breaks.Reported == "all" {
		rangeString = ">=5.2.0" + " <=" + PhpLatest
	} else {
		rangeString = ">=" + compat.Breaks.Low + " <=" + compat.Breaks.High
	}

	failRange, _ := semver.ParseRange(rangeString)

	for majorMinor, item := range phpVersions {

		if failRange(semver.MustParse(item["max"])) {
			broken = append(broken, majorMinor)
		}
	}

	sort.Strings(broken)

	return broken
}

// PhpMajorVersions returns only the major.minor parts from the `versions` variable as slice of strings.
func PhpMajorVersions() []string {
	versions := []string{}

	for key, _ := range phpVersions {
		versions = append(versions, key)
	}

	sort.Strings(versions)

	return versions
}

// MergeVersions takes slices of versions and returns a slice with unique values.
func MergeVersions(n ...[]string) []string {
	merged := []string{}

	for _, slice := range n {
		for _, value := range slice {
			if ! contains(merged, value) {
				merged = append(merged, value)
			}
		}
	}

	return merged
}

// ExcludeVersions removes the excluded versions from the given versions.
func ExcludeVersions(versions, exclude []string) []string {
	included := []string{}

	for _, version := range versions {
		if ! contains(exclude, version) && ! contains(included, version) {
			included = append(included, version)
		}
	}

	sort.Strings(included)
	return included
}

func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}