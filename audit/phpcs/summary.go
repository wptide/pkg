// phpcs package assumes that PHPCS is installed on the machine running these processors.
package phpcs

import (
	"io"
	"io/ioutil"
	"github.com/wptide/pkg/audit"
	"github.com/wptide/pkg/message"
	"github.com/wptide/pkg/tide"
	"encoding/json"
)

// PhpcsSummary implements audit.PostProcessor.
type PhpcsSummary struct {
	Report        io.Reader
	ParentProcess audit.Processor
}

// Kind is required to implement audit.PostProcessor.
func (p PhpcsSummary) Kind() string {
	return "phpcs_summary"
}

// Process runs the PHPCS summary post processing. It removes the detail from a PHPCS report.
// Required to implement audit.PostProcessor.
func (p *PhpcsSummary) Process(msg message.Message, result *audit.Result) {
	r := *result

	var byteSummary []byte
	var err error

	if p.Report != nil {
		byteSummary, err = ioutil.ReadAll(p.Report)
		if err != nil {
			return
		}
	}

	emptyResult := tide.AuditResult{}
	auditResults, ok := r[p.ParentProcess.Kind()].(*tide.AuditResult)

	// If we can't get the results there is nothing to do.
	if ! ok {
		return
	}

	// We don't want to override, so only add a summary if there is no summary.
	if auditResults.Summary == emptyResult.Summary {

		var fullResults tide.PhpcsResults
		err := json.Unmarshal(byteSummary, &fullResults)
		if err != nil {
			auditResults.Error += "\n" + "could not get phpcs results"
			return
		}

		summary := getPhpcsSummary(fullResults)

		auditResults.Summary = &tide.AuditSummary{
			PhpcsSummary: summary,
		}
	}
}

// getPhpcsSummary loops through all reported files and leaves only summary information.
func getPhpcsSummary(fullResults tide.PhpcsResults) *tide.PhpcsSummary {
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

// SetReport gets an io.Reader from the parent process. Required to implement audit.PostProcessor.
func (p *PhpcsSummary) SetReport(report io.Reader) {
	p.Report = report
}

// Parent gets a pointer to the parent process. Required to implement audit.PostProcessor.
func (p *PhpcsSummary) Parent(parent audit.Processor) {
	p.ParentProcess = parent
}
