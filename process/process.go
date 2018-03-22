// process is a package for managing a processing pipeline.
// Copied from https://github.com/rheinardkorf/go-concurrency/blob/master/11_pipeline_complex/pipe/process.go
package process

import (
	"context"
	"errors"
	"github.com/wptide/pkg/message"
	"io/ioutil"
	"os/exec"
	"os"
)

var (
	// Using exec.Command as a variable so that we can mock it in tests.
	execCommand = exec.Command

	// Using ioutil.WriteFile as a variable so that we can mock it in tests.
	writeFile = ioutil.WriteFile

	// Using os.Open as a variable so that we can mock it in tests.
	fileOpen = os.Open
)

// All our processes will use this as a base.
type Process struct {
	context   context.Context
	Message   message.Message        // Keeps track of the original message.
	Result    map[string]interface{} // Passes along a Result object.
	FilesPath string                 // Path of files to audit.
}

// A default implementation with an error nag. Not required, but serves as an example.
func (p *Process) Run() (<-chan error, error) {
	return nil, errors.New("process needs to implement Run()")
}

// We set our context so that it can be used in the processes to terminate goroutines
// if it needs to.
func (p *Process) SetContext(ctx context.Context) {
	p.context = ctx
}

func (p Process) Error(msg string) error {
	return errors.New(p.Message.Title + ": " + msg)
}

func (p *Process) SetMessage(msg message.Message) {
	p.Message = msg
}

func (p Process) GetMessage() message.Message {
	return p.Message
}

func (p *Process) SetResults(res map[string]interface{}) {
	p.Result = res
}

func (p Process) GetResult() map[string]interface{} {
	return p.Result
}

func (p *Process) SetFilesPath(path string) {
	p.FilesPath = path
}

func (p Process) GetFilesPath() string {
	return p.FilesPath
}

func (p *Process) CopyFields(proc Processor) {
	p.SetMessage(proc.GetMessage())
	p.SetResults(proc.GetResult())
	p.SetFilesPath(proc.GetFilesPath())
}

// Any process needs to implement the Processor interface.
type Processor interface {
	Run() (<-chan error, error)
	SetContext(ctx context.Context)
	SetMessage(msg message.Message)
	GetMessage() message.Message
	SetResults(res map[string]interface{})
	GetResult() map[string]interface{}
	SetFilesPath(path string)
	GetFilesPath() string
}
