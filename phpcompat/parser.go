package phpcompat

import (
	"github.com/wptide/pkg/tide"
	"strings"
	"regexp"
	"errors"
)

var (
	// Note: The order of these verbs are important.
	// There are many other phrases used in messages, but these are all we need
	// to assert the compatibility of a sniff violation.
	verbs = []string{
		"supported",
		"not present",   // or earlier
		"soft reserved", // "as of" "reserved keyword as of" , works like "deprecated"
		"reserved",      // "since", "introduced", "as of"
		"deprecated",    // optional "removed"
		"removed in",    // grammar added for PHP 7
		"removed",
		"prior to",
		"or earlier",
		"and earlier",
		"and lower",
		"<",
		"magic method",
		"available since",
		"since",
	}
)

// Parse takes a tide.PhpcsFilesMessage message and returns a Compatibility struct.
// It parses using the above verbs.
func Parse(e tide.PhpcsFilesMessage) (Compatibility, error) {

	versions := getVersions(e.Message)
	var breaks *CompatibilityRange
	var warns *CompatibilityRange

	if strings.ToLower(e.Type) == "error" {
		low, high, majorMinor, reported := GetVersionParts(versions[0], "")
		breaks = &CompatibilityRange{
			low,
			high,
			reported,
			majorMinor,
		}
	} else {
		low, high, majorMinor, reported := GetVersionParts(versions[0], "")
		warns = &CompatibilityRange{
			low,
			high,
			reported,
			majorMinor,
		}
	}

	matchList := strings.Join(verbs, "|")

	var re = regexp.MustCompile(`(?i)` + matchList + ``)
	matches := re.FindAllString(e.Message, -1)

	if len(matches) > 0 {

		matches = orderMatches(matches)

		// NOTE: Order is VERY important
		switch matches[0] {
		case "supported":
			fallthrough
		case "not present":
			low, high, majorMinor, reported := GetVersionParts(versions[0], "5.2.0")

			breaks = &CompatibilityRange{
				low,
				high,
				reported,
				majorMinor,
			}
		case "soft reserved":
			fallthrough
		case "deprecated":

			low, _, majorMinor, reported := GetVersionParts(versions[0], versions[0])
			warns = &CompatibilityRange{
				low,
				PhpLatest,
				reported,
				majorMinor,
			}

			if len(versions) > 1 {
				low, _, majorMinor, reported := GetVersionParts(versions[1], "")
				breaks = &CompatibilityRange{
					low,
					PhpLatest,
					reported,
					majorMinor,
				}

				warns.High = PreviousVersion(breaks.Low)
			}
		case "reserved":
			// We don't want to pick up "soft reserved".
			if len(matches) == 1 || (len(matches) > 1 && matches[0] != matches[1]) {

				low, _, majorMinor, reported := GetVersionParts(versions[0], versions[0])

				breaks = &CompatibilityRange{
					low,
					PhpLatest,
					reported,
					majorMinor,
				}
			}
		case "removed in":
			fallthrough
		case "removed":
			// We don't want to pick up "soft reserved" or "deprecated".
			if len(versions) == 1 {

				low, _, majorMinor, reported := GetVersionParts(versions[0], versions[0])

				breaks = &CompatibilityRange{
					low,
					PhpLatest,
					reported,
					majorMinor,
				}
			}
		case "prior to":

			// Annoying bad grammar for this error...
			if e.Source == "PHPCompatibility.PHP.EmptyNonVariable.Found" {
				version := PreviousVersion(versions[0])
				low, high, majorMinor, reported := GetVersionParts(version, "5.2.0")

				breaks = &CompatibilityRange{
					low,
					high,
					reported,
					majorMinor,
				}
			} else {
				low, high, majorMinor, reported := GetVersionParts(versions[0], versions[0])

				breaks = &CompatibilityRange{
					low,
					high,
					reported,
					majorMinor,
				}
			}
		case "<":
			// Another annoying grammar.
			if e.Source == "PHPCompatibility.PHP.TernaryOperators.MiddleMissing" {
				low, high, majorMinor, reported := GetVersionParts(versions[0], versions[0])

				breaks = &CompatibilityRange{
					low,
					high,
					reported,
					majorMinor,
				}
			}
		case "or earlier":
			fallthrough
		case "and earlier":
			fallthrough
		case "and lower":
			low, high, majorMinor, reported := GetVersionParts(versions[0], "5.2.0")

			breaks = &CompatibilityRange{
				low,
				high,
				reported,
				majorMinor,
			}

		case "magic method":
			low, high, majorMinor, reported := GetVersionParts("all", "")

			breaks = &CompatibilityRange{
				low,
				high,
				reported,
				majorMinor,
			}

		case "available since":
			if breaks != nil {
				breaks.Low = "5.2.1"
				breaks.High = PreviousVersion(breaks.MajorMinor)
			}
			if warns != nil {
				warns.Low = "5.2.1"
				warns.High = PreviousVersion(warns.MajorMinor)
			}

		case "since":
			if breaks != nil {
				breaks.High = PhpLatest
			}
			if warns != nil {
				warns.High = PhpLatest
			}
		}
	} else {
		return Compatibility{}, errors.New("Could not parse message.")
	}

	return Compatibility{
		e.Source,
		breaks,
		warns,
	}, nil
}

func getVersions(line string) []string {
	pattern := `(?i)((\d+\.)+\d+)|(\ball\b)`
	var re = regexp.MustCompile(pattern)
	result := re.FindAllString(line, -1)

	if len(result) == 0 {
		// Becaise they don't like minors?
		pattern := `(?i)PHP 7`
		var re = regexp.MustCompile(pattern)
		result := re.FindAllString(line, -1)

		if len(result) == 0 {
			return []string{"all"}
		}

		return []string{"7.0"}
	}

	return result
}

func orderMatches(m []string) (matches []string) {

	found := func(value string, matches []string) bool {
		for _, val := range matches {
			if val == value {
				return true
			}
		}
		return false
	}

	for _, verb := range verbs {

		for _, match := range m {

			if strings.ToLower(verb) == strings.ToLower(match) && ! found(match, matches) {
				matches = append(matches, strings.ToLower(match))
			}
		}
	}

	return;
}
