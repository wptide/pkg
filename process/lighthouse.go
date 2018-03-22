package process

import (
	"errors"
	"strings"
	"github.com/wptide/pkg/log"
	"github.com/wptide/pkg/tide"
	"github.com/wptide/pkg/storage"
	"fmt"
	"encoding/json"
	"github.com/wptide/pkg/shell"
)

var runner shell.Runner

// Ingest defines the structure for our Ingest process.
type Lighthouse struct {
	Process                                 // Inherits methods from Process.
	In              <-chan Processor        // Expects a processor channel as input.
	Out             chan Processor          // Send results to an output channel.
	TempFolder      string                  // Path to a temp folder where reports will be generated.
	StorageProvider storage.StorageProvider // Storage provider to upload reports to.
}

func (lh *Lighthouse) Run() (<-chan error, error) {

	if lh.TempFolder == "" {
		return nil, errors.New("no temp folder provided for lighthouse reports")
	}

	if lh.StorageProvider == nil {
		return nil, errors.New("no storage provider for lighthouse reports")
	}

	if lh.In == nil {
		return nil, errors.New("requires a previous process")
	}
	if lh.Out == nil {
		return nil, errors.New("requires a next process")
	}

	errc := make(chan error, 1)

	go func() {
		defer close(errc)
		for {
			select {
			case in := <-lh.In:

				// Copy Process fields from `in` process.
				lh.CopyFields(in)

				// Assume that the rest of the message is also broken.
				// Don't pass this down the pipe.
				if lh.Message.Title == "" {
					errc <- lh.Error("invalid message")
					break
				}

				// Run the process.
				// If processing produces an error send it up the error channel.
				for _, audit := range *lh.Message.Audits {
					if audit.Type == "lighthouse" {
						if err := lh.process(); err != nil {
							// Pass the error up the error channel.
							errc <- err
							// Don't break, the message is still useful to other processes.
						}
					}
				}

				// Send process to the out channel.
				lh.Out <- lh
			case <-lh.context.Done():
				return
			}
		}

	}()

	return errc, nil
}

func (lh *Lighthouse) process() error {
	log.Log(lh.Message.Title, "Running Lighthouse Audit...")

	var results *tide.LighthouseSummary

	// Note: This assumes the shell script `lh` is in $PATH and contains the following command:
	// `lighthouse --quiet --chrome-flags="--headless --disable-gpu --no-sandbox" --output=json --output-path=stdout $@`
	cmdName := "lh"
	cmdArgs := []string{fmt.Sprintf("https://wp-themes.com/%s", lh.Message.Slug)}

	// Prepare the command and set the stdOut pipe.
	resultBytes, errorBytes, err, _ := runner.Run(cmdName, cmdArgs...)

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
	fullResults, err := lh.uploadToStorage(resultBytes)
	if err != nil {
		return err
	}

	if fullResults != nil {
		auditResult.Full = fullResults.Full
		auditResult.Details.Type = fullResults.Full.Type
		auditResult.Details.Key = fullResults.Full.Key
		auditResult.Details.BucketName = fullResults.Full.BucketName
	}

	auditResult.Summary = tide.AuditSummary{
		LighthouseSummary: results,
	}

	lh.Result["lighthouse"] = auditResult

	log.Log(lh.Message.Title, "Lighthouse process complete.")

	return nil
}

func (lh Lighthouse) uploadToStorage(buffer []byte) (*tide.AuditResult, error) {

	var results *tide.AuditResult

	checksum, checksumOk := lh.Result["checksum"].(string)
	if ! checksumOk {
		return nil, errors.New("there was no checksum to be used for filenames")
	}

	storageRef := checksum + "-lighthouse-full.json"
	filename := strings.TrimRight(lh.TempFolder, "/") + "/" + storageRef

	err := writeFile(filename, buffer, 0644)
	if err != nil {
		return nil, errors.New("could not write lighthouse audit to tempFolder")
	}

	err = lh.StorageProvider.UploadFile(filename, storageRef)

	if err == nil {
		results = &tide.AuditResult{
			Full: tide.AuditDetails{
				Type:       lh.StorageProvider.Kind(),
				Key:        storageRef,
				BucketName: lh.StorageProvider.CollectionRef(),
			},
		}
	}

	return results, err
}
