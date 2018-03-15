package process

import (
	"errors"
)

// Ingest defines the structure for our Ingest process.
type Tide struct {
	Process              // Inherits methods from Process.
	In  <-chan Processor // Expects a processor channel as input.
	Out chan Processor   // Send results to an output channel.
}

func (tc *Tide) Run() (<-chan error, error) {

	if tc.In == nil {
		return nil, errors.New("requires a previous process")
	}
	if tc.Out == nil {
		return nil, errors.New("requires a next process")
	}

	errc := make(chan error, 1)

	go func() {
		defer close(errc)
		for {
			select {
			case in := <-tc.In:

				// Copy Process fields from `in` process.
				tc.CopyFields(in)

				// Run the process.
				// If processing produces an error send it up the error channel.
				if err := tc.process(); err != nil {
					// Pass the error up the error channel.
					errc <- err
					break
				}

				// Send process to the out channel.
				tc.Out <- tc
			case <-tc.context.Done():
				return
			}
		}

	}()

	return errc, nil
}

func (tc *Tide) process() error {

	return nil
}
