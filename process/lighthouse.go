package process

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/wptide/pkg/log"
	"github.com/wptide/pkg/shell"
	"github.com/wptide/pkg/storage"
	"github.com/wptide/pkg/tide"
)

var (
	lhRunner      shell.Runner
	defaultRunner shell.Runner = &shell.Command{}
)

// Lighthouse defines the structure for our Lighthouse process.
type Lighthouse struct {
	Process                          // Inherits methods from Process.
	In              <-chan Processor // Expects a processor channel as input.
	Out             chan Processor   // Send results to an output channel.
	TempFolder      string           // Path to a temp folder where reports will be generated.
	StorageProvider storage.Provider // Storage provider to upload reports to.
}

// Run runs the process in a pipeline.
func (lh *Lighthouse) Run(errc *chan error) error {
	if lh.TempFolder == "" {
		return errors.New("no temp folder provided for lighthouse reports")
	}

	if lh.StorageProvider == nil {
		return errors.New("no storage provider for lighthouse reports")
	}

	if lh.In == nil {
		return errors.New("requires a previous process")
	}
	if lh.Out == nil {
		return errors.New("requires a next process")
	}

	go func() {
		for {
			select {
			case in := <-lh.In:

				// Copy Process fields from `in` process.
				lh.CopyFields(in)

				// Assume that the rest of the message is also broken.
				// Don't pass this down the pipe.
				if lh.Message.Title == "" {
					*errc <- errors.New("Lighthouse Error: " + lh.Error("invalid message").Error())
					continue
				}

				// Run the process.
				// If processing produces an error send it up the error channel.
				for _, audit := range lh.Message.Audits {
					if audit.Type == "lighthouse" {
						if err := lh.Do(); err != nil {
							// Pass the error up the error channel.
							*errc <- errors.New("Lighthouse Error: " + err.Error())
							// Don't break, the message is still useful to other processes.
						}
					}
				}

				// Send process to the out channel.
				lh.Out <- lh
			}
		}

	}()

	return nil
}

// Do executes the process.
func (lh *Lighthouse) Do() error {
	log.Log(lh.Message.Title, "Running Lighthouse Audit...")

	if lhRunner == nil {
		lhRunner = defaultRunner
	}

	var results *tide.LighthouseSummary

	// Note: This assumes the shell script `lh` is in $PATH and contains the following command:
	// `lighthouse --quiet --chrome-flags="--headless --disable-gpu --no-sandbox" --output=json --output-path=stdout $@`
	cmdName := "lh"
	cmdArgs := []string{fmt.Sprintf("https://wp-themes.com/%s", lh.Message.Slug)}

	// Prepare the command and set the stdOut pipe.
	resultBytes, errorBytes, _, err := lhRunner.Run(cmdName, cmdArgs...)

	if len(errorBytes) > 0 {
		return lh.Error("lighthouse command failed: " + string(errorBytes))
	}

	// Unmarshal the body response into a LightHouseReport object.
	err = json.Unmarshal(resultBytes, &results)
	if err != nil {
		return err
	}

	auditResult := tide.AuditResult{}

	// Upload and get full results.
	log.Log(lh.Message.Title, "Uploading results to remote storage.")
	rawResults, err := lh.uploadToStorage(resultBytes)
	if err != nil {
		return err
	}

	if rawResults != nil {
		auditResult.Raw = rawResults.Raw
	}

	auditResult.Summary = tide.AuditSummary{
		LighthouseSummary: results,
	}

	result := *lh.Result
	result["lighthouse"] = auditResult
	lh.Result = &result

	log.Log(lh.Message.Title, "Lighthouse process complete.")

	return nil
}

func (lh Lighthouse) uploadToStorage(buffer []byte) (*tide.AuditResult, error) {

	var results *tide.AuditResult

	result := *lh.Result
	checksum, checksumOk := result["checksum"].(string)
	if !checksumOk {
		return nil, errors.New("there was no checksum to be used for filenames")
	}

	storageRef := checksum + "-lighthouse-raw.json"
	filename := strings.TrimRight(lh.TempFolder, "/") + "/" + storageRef

	err := writeFile(filename, buffer, 0644)
	if err != nil {
		return nil, errors.New("could not write lighthouse audit to tempFolder")
	}

	err = lh.StorageProvider.UploadFile(filename, storageRef)

	if err == nil {
		results = &tide.AuditResult{
			Raw: tide.AuditDetails{
				Type:     lh.StorageProvider.Kind(),
				FileName: storageRef,
				Path:     lh.StorageProvider.CollectionRef(),
			},
		}
	}

	return results, err
}
