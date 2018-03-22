package phpcs

import "github.com/wptide/pkg/tide"

// getPhpcsSummary loops through all reported files and leaves only summary information.
func GetPhpcsSummary(fullResults tide.PhpcsResults) *tide.PhpcsSummary {
	summary := &tide.PhpcsSummary{
		ErrorsCount:   fullResults.Totals.Errors,
		WarningsCount: fullResults.Totals.Warnings,
		FilesCount:    len(fullResults.Files),
		Files: make(map[string]struct {
			Errors   int `json:"errors"`
			Warnings int `json:"warnings"`
		}),
	}

	// Iterate files and only get summary data.
	for filename, data := range fullResults.Files {
		summary.Files[filename] = struct {
			Errors   int `json:"errors"`
			Warnings int `json:"warnings"`
		}{
			data.Errors,
			data.Warnings,
		}
	}
	return summary
}