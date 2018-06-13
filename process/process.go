// Package process is a package for managing a processing pipeline.
// Copied from https://github.com/rheinardkorf/go-concurrency/blob/master/11_pipeline_complex/pipe/process.go
package process

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/wptide/pkg/message"
)

var (
	// Using exec.Command as a variable so that we can mock it in tests.
	execCommand = exec.Command

	// Using ioutil.WriteFile as a variable so that we can mock it in tests.
	writeFile = ioutil.WriteFile

	// Using os.Open as a variable so that we can mock it in tests.
	fileOpen = os.Open
)

// Result is an interface map of the processed results.
type Result map[string]interface{}

// Process is the base for all processes.
type Process struct {
	context   context.Context
	Message   message.Message // Keeps track of the original message.
	Result    *Result         // Passes along a Result object.
	FilesPath string          // Path of files to audit.
}

// Run is a default implementation with an error nag. Not required, but serves as an example.
func (p *Process) Run() (<-chan error, error) {
	return nil, errors.New("process needs to implement Run()")
}

// SetContext sets the context so that it can be used in the processes to terminate goroutines
// if it needs to.
func (p *Process) SetContext(ctx context.Context) {
	p.context = ctx
}

// Error returns a new process error.
func (p Process) Error(msg string) error {
	return errors.New(p.Message.Title + ": " + msg)
}

// SetMessage is used to set the Message for this process (used for copying the message).
func (p *Process) SetMessage(msg message.Message) {
	p.Message = msg
}

// GetMessage returns the message object for this process.
func (p Process) GetMessage() message.Message {
	return p.Message
}

// SetResults sets the results for this process (used for copying results from process to process).
func (p *Process) SetResults(res *Result) {
	p.Result = res
}

// GetResult gets the results attached to this process.
func (p Process) GetResult() *Result {
	return p.Result
}

// SetFilesPath sets the path of the code to run the process againsts.
func (p *Process) SetFilesPath(path string) {
	p.FilesPath = path
}

// GetFilesPath gets the path of the source used for processing.
func (p Process) GetFilesPath() string {
	return p.FilesPath
}

// CopyFields copies required fields from one process to another.
func (p *Process) CopyFields(proc Processor) {
	p.SetMessage(proc.GetMessage())
	p.SetResults(proc.GetResult())
	p.SetFilesPath(proc.GetFilesPath())
}

// Processor is an interface for all processors.
type Processor interface {
	Run(*chan error) error
	Do() error
	SetContext(ctx context.Context)
	SetMessage(msg message.Message)
	GetMessage() message.Message
	SetResults(res *Result)
	GetResult() *Result
	SetFilesPath(path string)
	GetFilesPath() string
}
