// phpcs package assumes that PHPCS is installed on the machine running these processors.
package phpcs

import (
	"io"
	"io/ioutil"
	"github.com/wptide/pkg/audit"
	"github.com/wptide/pkg/message"
	"github.com/wptide/pkg/tide"
	"encoding/json"
	"strings"
	"errors"
	"github.com/wptide/pkg/storage"
	"github.com/wptide/pkg/phpcompat"
)

var (
	// Making this overridable so that it can be tested.
	writeFile = ioutil.WriteFile
)

// PhpcsSummary implements audit.PostProcessor.
type PhpCompat struct {
	Report        io.Reader
	ParentProcess audit.Processor
	resultPath    string
	resultFile    string
}

// Kind is required to implement audit.PostProcessor.
func (p PhpCompat) Kind() string {
	return "phpcs_phpcompatibility"
}

type PhpCompatDetails struct {
	Totals   map[string]int                       `json:"totals"`
	ErrorMap map[string][]string                  `json:"error_map"`
	Errors   map[string]PhpCompatDetailsViolation `json:"errors"`
}

type PhpCompatDetailsViolation struct {
	Message  string                    `json:"message"`
	Source   string                    `json:"source"`
	Type     string                    `json:"type"`
	Severity int                       `json:"severity"`
	Versions []string                  `json:"versions"`
	Files    map[string][]FilePosition `json:"files"`
}

type FilePosition struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Process runs the PHPCompatiblity post processing and determines the compatible versions.
//
// A detailed report is also sent to a storage provider which contains a structure ordered by each violating sniff which gives:
// - the PHP versions effected by the violation
// - the impacted files and relevant phpcs messages
//
// Process is required to implement audit.PostProcessor.
func (p *PhpCompat) Process(msg message.Message, result *audit.Result) {
	r := *result

	// Byte buffer for reading from the report file.
	var byteSummary []byte
	var err error

	if p.Report != nil {
		// Read the report into the byte buffer.
		byteSummary, _ = ioutil.ReadAll(p.Report)
	}

	// We need an empty result to compare nil values.
	emptyResult := tide.AuditResult{}

	// Get current audit results from parent.
	auditResults, ok := r[p.ParentProcess.Kind()].(*tide.AuditResult)

	// If we can't get the results there is nothing to do.
	if ! ok {
		errMsg := "could not get results from parent process"
		r[p.Kind()+"Error"] = errors.New(errMsg)
		return
	}

	// Full "raw" results from the PHPCS process.
	var fullResults tide.PhpcsResults
	err = json.Unmarshal(byteSummary, &fullResults)
	if err != nil {
		errMsg := "could not get phpcs results"
		r[p.Kind()+"Error"] = errors.New(errMsg)
		auditResults.Error += "\n" + errMsg
		return
	}

	brokenVersions := []string{}

	// Dynamically creating our struct for JSON output.
	details := &PhpCompatDetails{
		Totals: map[string]int{
			"errors":   fullResults.Totals.Errors,
			"warnings": fullResults.Totals.Warnings,
		},
		ErrorMap: make(map[string][]string),
		Errors:   make(map[string]PhpCompatDetailsViolation),
	}

	// Iterate files and only get summary data.
	for filename, data := range fullResults.Files {
		for _, sniffMessage := range data.Messages {

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
		}
	}

	// Will Marshall without error because the sources map gets initialized with `make()`
	res, _ := json.Marshal(details)

	// Determines where the detailed result will be written.
	p.resultFile, p.resultPath, err = p.reportPath(r, "detail")

	// Attempt to write results to disk.
	err = writeFile(p.resultPath, res, 0644)
	if err != nil {
		errMsg := "could not write PHPCompatibility details to disk"
		r[p.Kind()+"Error"] = errors.New(errMsg)
		auditResults.Error += "\n" + errMsg
		return
	}

	// Attempt to upload the report file to a storage provider.
	auditDetails, err := p.uploadResults(r)
	if err != nil {
		errMsg := "could not write PHPCompatibility details to file store"
		r[p.Kind()+"Error"] = errors.New(errMsg)
		auditResults.Error += "\n" + errMsg
		return
	}

	// We don't want to override, so only add the details if there are no details.
	if auditResults.Details == emptyResult.Details {
		auditResults.Details = auditDetails.Details
	}

	// Remove the broken versions from the PHP major versions to get the compatible versions.
	auditResults.CompatibleVersions = phpcompat.ExcludeVersions(phpcompat.PhpMajorVersions(), brokenVersions)
}

// SetReport gets an io.Reader from the parent process. Required to implement audit.PostProcessor.
func (p *PhpCompat) SetReport(report io.Reader) {
	p.Report = report
}

// Parent gets a pointer to the parent process. Required to implement audit.PostProcessor.
func (p *PhpCompat) Parent(parent audit.Processor) {
	p.ParentProcess = parent
}

// uploadResults attempts to upload the results to a storage provider "fileStore"
// and returns a tide.AuditResult with "Details" referencing the storage provider.
func (p PhpCompat) uploadResults(r audit.Result) (results *tide.AuditResult, err error) {

	var storageProvider *storage.StorageProvider
	var ok bool

	if p.resultFile == "" || p.resultPath == "" {
		err = errors.New("no result path given")
		return
	}

	if storageProvider, ok = r["fileStore"].(*storage.StorageProvider); !ok {
		err = errors.New("could not get fileStore to upload to")
		return
	}

	sP := *storageProvider
	err = sP.UploadFile(p.resultPath, p.resultFile)

	if err == nil {
		results = &tide.AuditResult{
			Details: struct {
				Type       string `json:"type,omitempty"`
				Key        string `json:"key,omitempty"`
				BucketName string `json:"bucket_name,omitempty"`
				*tide.PhpcsResults
				*tide.LighthouseResults
			}{
				Type:       sP.Kind(),
				Key:        p.resultFile,
				BucketName: sP.CollectionRef(),
			},
		}
	}

	return
}

// reportPath generates a filename and destination path.
func (p PhpCompat) reportPath(r audit.Result, fileSuffix string) (filename, path string, err error) {

	var checksum, tempFolder string
	var ok bool

	if tempFolder, ok = r["tempFolder"].(string); ! ok {
		err = errors.New("no tempFolder to write results to before upload to fileStore")
		return
	}

	if checksum, ok = r["checksum"].(string); ! ok {
		err = errors.New("there was no checksum to be used for filenames")
		return
	}

	filename = checksum + "-" + p.Kind()
	if fileSuffix != "" {
		filename += "-" + fileSuffix + ".json"
	} else {
		filename += ".json"
	}

	path = strings.TrimRight(tempFolder, "/") + "/" + filename

	return
}
