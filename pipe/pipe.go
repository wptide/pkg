// Package pipe is adapted from https://github.com/rheinardkorf/go-concurrency/blob/master/11_pipeline_complex/pipe/pipe.go
package pipe

import (
	"context"
	"errors"

	"github.com/wptide/pkg/process"
)

// Pipe represents a pipe that contains multiple processes.
type Pipe struct {
	processes  []process.Processor
	errors     []<-chan error
	context    context.Context
	cancelFunc context.CancelFunc
}

// New creates a new Pipe and then runs the init() method which sets a cancelable context.
func New() *Pipe {
	p := &Pipe{}
	p.init()
	return p
}

// WithProcesses creates a new Pipe, but accepts a slice of Processors to run later.
func WithProcesses(procs ...process.Processor) *Pipe {
	p := New()
	p.AddProcesses(procs...)
	return p
}

// init gets a context and sets the cancelFunction.
func (p *Pipe) init() {
	p.context, p.cancelFunc = context.WithCancel(context.Background())
}

// AddProcess adds a single process to the processes slice.
func (p *Pipe) AddProcess(proc process.Processor) error {
	if proc == nil {
		return errors.New("could not add nil processor")
	}

	// Set the context for our process.
	// This is used inside the processes to look for a cancel message from the context.
	proc.SetContext(p.context)

	p.processes = append(p.processes, proc)
	return nil
}

// AddProcesses adds a multiple processes to the processes slice.
func (p *Pipe) AddProcesses(procs ...process.Processor) error {
	for _, proc := range procs {
		err := p.AddProcess(proc)
		if err != nil {
			return errors.New("some processors could not be added")
		}
	}
	return nil
}

// Run iterates over the processes slice and starts each process.
func (p *Pipe) Run(errc *chan error) error {
	defer p.cancelFunc()

	for _, proc := range p.processes {
		err := proc.Run(errc)
		if err != nil {
			return err
		}
	}

	return nil
}
