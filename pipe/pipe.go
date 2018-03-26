// Adapted from https://github.com/rheinardkorf/go-concurrency/blob/master/11_pipeline_complex/pipe/pipe.go
package pipe

import (
	"context"
	"errors"
	"sync"
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

// AddProcess adds a multiple processes to the processes slice.
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
func (p *Pipe) Run() error {
	defer p.cancelFunc()

	for _, proc := range p.processes {
		errc, err := proc.Run()
		if err != nil {
			return err
		}
		p.errors = append(p.errors, errc)
	}

	return p.wait()
}

// wait ranges over all the error channels which causes the pipe to block until all processes are completed.
// Based on https://medium.com/statuscode/pipeline-patterns-in-go-a37bb3a7e61d.
func (p Pipe) wait() error {
	errc := p.mergeErrors()

	for err := range errc {
		if err != nil {
			return err
		}
	}

	return nil
}

// mergeErrors merges multiple channels of errors.
// Based on https://blog.golang.org/pipelines.
// Based on https://medium.com/statuscode/pipeline-patterns-in-go-a37bb3a7e61d.
func (p Pipe) mergeErrors() <-chan error {
	var wg sync.WaitGroup

	out := make(chan error, len(p.errors))

	// Create output function.
	output := func(ce <-chan error) {

		if len(ce) > 0 {
			for err := range ce {
				out <- err
			}
		}
		wg.Done()
	}

	wg.Add(len(p.errors))
	for _, err := range p.errors {
		go output(err)
	}

	// Drain the error channel.
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
