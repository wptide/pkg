package phpcs

import (
	"io/ioutil"

	"github.com/wptide/pkg/phpcompat"
	"github.com/wptide/pkg/tide"
)

var (
	writeFile = ioutil.WriteFile
)

// PhpCompatDetails describes details from `phpcs` phpcompatibility rules.
type PhpCompatDetails struct {
	Totals     map[string]int                       `json:"totals"`
	ErrorMap   map[string][]string                  `json:"error_map"`
	WarningMap map[string][]string                  `json:"warning_map"`
	Errors     map[string]PhpCompatDetailsViolation `json:"errors"`
	Warnings   map[string]PhpCompatDetailsViolation `json:"warnings"`
}

// PhpCompatDetailsViolation describes a single violation.
type PhpCompatDetailsViolation struct {
	Message  string                    `json:"message"`
	Source   string                    `json:"source"`
	Type     string                    `json:"type"`
	Severity int                       `json:"severity"`
	Versions []string                  `json:"versions"`
	Files    map[string][]FilePosition `json:"files"`
}

// FilePosition describes where a violation occurred.
type FilePosition struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// GetPhpcsCompatibility runs the PHPCompatiblity post processing and determines the compatible versions.
//
// A detailed report is also sent to a storage provider which contains a structure ordered by each violating sniff which gives:
// - the PHP versions effected by the violation
// - the impacted files and relevant phpcs messages
//
// Process is required to implement audit.PostProcessor.
func GetPhpcsCompatibility(fullResults tide.PhpcsResults) ([]string, interface{}) {

	brokenVersions := []string{}

	// Dynamically creating our struct for JSON output.
	details := &PhpCompatDetails{
		Totals: map[string]int{
			"errors":   fullResults.Totals.Errors,
			"warnings": fullResults.Totals.Warnings,
		},
		ErrorMap:   make(map[string][]string),
		WarningMap: make(map[string][]string),
		Errors:     make(map[string]PhpCompatDetailsViolation),
		Warnings:   make(map[string]PhpCompatDetailsViolation),
	}

	// Iterate files and only get summary data.
	for filename, data := range fullResults.Files {
		for _, sniffMessage := range data.Messages {

			if sniffMessage.Type == "ERROR" {
				// Create the new Violation if we don't have it already.
				// This happens only once because we group failures.
				if _, ok := details.Errors[sniffMessage.Source]; !ok {
					// Create the object.
					violation := PhpCompatDetailsViolation{
						Message:  sniffMessage.Message,
						Source:   sniffMessage.Source,
						Type:     sniffMessage.Type,
						Severity: sniffMessage.Severity,
						Files:    make(map[string][]FilePosition),
					}

					// Get incompatible versions
					broken := phpcompat.BreaksVersions(sniffMessage)
					violation.Versions = broken

					// Add the source to each broken version.
					for _, version := range broken {
						details.ErrorMap[version] = append(details.ErrorMap[version], sniffMessage.Source)
					}

					// Add to broken versions so that we can determine compatibility later.
					brokenVersions = phpcompat.MergeVersions(brokenVersions, broken)

					details.Errors[sniffMessage.Source] = violation
				}

				// Each violating file needs to be added to the particular violation.
				details.Errors[sniffMessage.Source].Files[filename] = append(
					details.Errors[sniffMessage.Source].Files[filename],
					FilePosition{
						sniffMessage.Line,
						sniffMessage.Column,
					},
				)
			} else if sniffMessage.Type == "WARNING"  {
				if _, ok := details.Warnings[sniffMessage.Source]; !ok {
					violation := PhpCompatDetailsViolation{
						Message:  sniffMessage.Message,
						Source:   sniffMessage.Source,
						Type:     sniffMessage.Type,
						Severity: sniffMessage.Severity,
						Files:    make(map[string][]FilePosition),
					}

					broken := phpcompat.NonBreakingVersions(sniffMessage)
					violation.Versions = broken

					for _, version := range broken {
						details.WarningMap[version] = append(details.WarningMap[version], sniffMessage.Source)
					}

					details.Warnings[sniffMessage.Source] = violation
				}

				details.Warnings[sniffMessage.Source].Files[filename] = append(
					details.Warnings[sniffMessage.Source].Files[filename],
					FilePosition{
						sniffMessage.Line,
						sniffMessage.Column,
					},
				)
			}
		}
	}

	compatibleVersion := phpcompat.ExcludeVersions(phpcompat.PhpMajorVersions(), brokenVersions)
	return compatibleVersion, details
}
