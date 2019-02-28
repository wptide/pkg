package process

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/wptide/pkg/log"
	"github.com/wptide/pkg/message"
	"github.com/wptide/pkg/process/phpcs"
	"github.com/wptide/pkg/shell"
	"github.com/wptide/pkg/storage"
	"github.com/wptide/pkg/tide"
)

var (
	phpcsRunner shell.Runner
)

// Phpcs defines the structure for our Phpcs process.
type Phpcs struct {
	Process                          // Inherits methods from Process.
	In              <-chan Processor // Expects a processor channel as input.
	Out             chan Processor   // Send results to an output channel.
	Config          Result           // Additional config.
	TempFolder      string           // Path to a temp folder where reports will be generated.
	StorageProvider storage.Provider // Storage provider to upload reports to.
}

// Run executes the process in a pipe.
func (cs *Phpcs) Run(errc *chan error) error {

	if cs.TempFolder == "" {
		return errors.New("no temp folder provided for phpcs reports")
	}

	if cs.StorageProvider == nil {
		return errors.New("no storage provider for phpcs reports")
	}

	if cs.In == nil {
		return errors.New("requires a previous process")
	}
	if cs.Out == nil {
		return errors.New("requires a next process")
	}

	go func() {
		for {
			select {
			case in := <-cs.In:

				// Copy Process fields from `in` process.
				cs.CopyFields(in)

				result := *cs.Result

				// Run the process.
				// If processing produces an error send it up the error channel.
				for _, audit := range cs.Message.Audits {
					if audit.Type == "phpcs" {
						result["phpcsCurrentAudit"] = audit
						cs.SetResults(&result)
						if err := cs.Do(); err != nil {
							// Pass the error up the error channel.
							*errc <- errors.New("PHPCS Error: " + err.Error())
							// Don't break, the message is still useful to other processes.
						}
					}
				}

				// Send process to the out channel.
				cs.Out <- cs
			}
		}

	}()

	return nil
}

// Do executes the process.
func (cs *Phpcs) Do() error {

	log.Log(cs.Message.Title, "Running PHPCS Audit...")

	if phpcsRunner == nil {
		phpcsRunner = defaultRunner
	}

	result := *cs.Result

	// Get the current audit from the result.
	audit := result["phpcsCurrentAudit"].(*message.Audit)

	// Try to get filesPath from results first.
	if path, ok := result["filesPath"].(string); ok {
		cs.SetFilesPath(path)
	}

	standard := audit.Options.Standard
	if standard == "" {
		return errors.New("could not determine standard for report")
	}

	checksum, ok := result["checksum"].(string)
	if !ok {
		return errors.New("could not determine checksum")
	}

	//return errors.New("could not determine files path")
	if cs.GetFilesPath() == "" {
		return errors.New("could not determine files path")
	}

	path := cs.GetFilesPath() + "/unzipped"

	kind := strings.ToLower(audit.Type) + "_" + strings.ToLower(standard)
	filename := checksum + "-" + kind + "-raw.json"
	pathPrefix := strings.TrimRight(cs.TempFolder, "/") + "/"
	filepath := pathPrefix + filename

	// Provide in implementation, not from message.
	parallel, ok := cs.Config["parallel"].(int)
	if !ok {
		parallel = 1
	}

	// Get encoding from message and provide a fallback.
	encoding := audit.Options.Encoding
	if encoding == "" {
		encoding = "utf-8"
	}

	cliStandard := standard
	if audit.Options.StandardOverride != "" {
		cliStandard = audit.Options.StandardOverride
	}

	cmdName := "phpcs"
	cmdArgs := []string{
		"--extensions=php",
		"--ignore=" + audit.Options.Ignore,
		"--standard=" + cliStandard,
		"--encoding=" + encoding,
		"--basepath=" + path, // Remove this part from the filenames in PHPCS report.
		"--report=json",
		"--report-json=" + filepath,
		"--parallel=" + strconv.Itoa(parallel),
		"-d",              // Required to be before "memory_limit".
		"memory_limit=128M", // Leave memory handling up to the system.
	}

	// @todo fix message to accept array of options.
	//for _, pair := range audit.Options.RuntimeSet {
	split := strings.Split(audit.Options.RuntimeSet, " ")
	if len(split) == 2 {
		cmdArgs = append(cmdArgs, "--runtime-set")
		cmdArgs = append(cmdArgs, split[0])
		cmdArgs = append(cmdArgs, split[1])
	}
	//}

	cmdArgs = append(cmdArgs, path)
	cmdArgs = append(cmdArgs, "-q")

	// Prepare the command and set the stdOut pipe.
	resultBytes, _, exitCode, err := phpcsRunner.Run(cmdName, cmdArgs...)
	if err != nil {
		return err
	}

	log.Log(cs.Message.Title, fmt.Sprintf("phpcs output:\n %s", strings.TrimSpace(string(resultBytes))))

	// We already have a reference to the report file, so lets upload and get the storage reference in a result.
	log.Log(cs.Message.Title, "Uploading "+standard+" results to remote storage.")

	fType, fFileName, fPath, err := cs.uploadToStorage(filepath, filename)
	if err != nil {
		return err
	}

	// Initialise the result and set the "Raw" entry to the uploaded file.
	auditResults := tide.AuditResult{
		Raw: tide.AuditDetails{
			Type:     fType,
			FileName: fFileName,
			Path:     fPath,
		},
	}

	// `uploadToStorage` already did the error checking.
	fileReader, _ := fileOpen(filepath)
	defer fileReader.Close()

	report, _ := ioutil.ReadAll(fileReader)

	var phpcsResults *tide.PhpcsResults
	err = json.Unmarshal(report, &phpcsResults)
	if err != nil {
		return err
	}

	// Get the PHPCS Summary.
	summary := phpcs.GetPhpcsSummary(*phpcsResults)
	auditResults.Summary = tide.AuditSummary{PhpcsSummary: summary}

	// Only PHPCompatibility provides parsed results.
	// @todo Abstract this later.
	if kind == "phpcs_phpcompatibility" {
		compatibleVersions, compatResults := phpcs.GetPhpcsCompatibility(*phpcsResults)

		resultsJSON, _ := json.Marshal(compatResults)

		fname := checksum + "-" + kind + "-parsed.json"
		fpath := pathPrefix + fname

		err = writeFile(fpath, resultsJSON, os.ModePerm)
		if err != nil {
			return err
		}

		fType, fFileName, fPath, err := cs.uploadToStorage(fpath, fname)
		if err != nil {
			return err
		}

		auditResults.Parsed = tide.AuditDetails{
			Type:     fType,
			FileName: fFileName,
			Path:     fPath,
		}

		auditResults.CompatibleVersions = compatibleVersions
	}

	// Reset current audit.
	result["phpcsCurrentAudit"] = nil

	result[kind] = auditResults
	cs.Result = &result

	log.Log(cs.Message.Title, fmt.Sprintf("phpcs (%s) process completed with exit code: %d\n", standard, exitCode))

	return nil
}

func (cs Phpcs) uploadToStorage(filepath, filename string) (fType, fFileName, fPath string, err error) {
	err = cs.StorageProvider.UploadFile(filepath, filename)

	if err == nil {
		fType = cs.StorageProvider.Kind()
		fFileName = filename
		fPath = cs.StorageProvider.CollectionRef()
	}

	return fType, fFileName, fPath, err
}
