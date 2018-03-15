package process

import (
	"errors"
)

// Ingest defines the structure for our Ingest process.
type Phpcs struct {
	Process              // Inherits methods from Process.
	In  <-chan Processor // Expects a processor channel as input.
	Out chan Processor   // Send results to an output channel.
}

func (cs *Phpcs) Run() (<-chan error, error) {

	if cs.In == nil {
		return nil, errors.New("requires a previous process")
	}
	if cs.Out == nil {
		return nil, errors.New("requires a next process")
	}

	errc := make(chan error, 1)

	go func() {
		defer close(errc)
		for {
			select {
			case in := <-cs.In:

				// Copy Process fields from `in` process.
				cs.CopyFields(in)

				// Run the process.
				// If processing produces an error send it up the error channel.
				if err := cs.process(); err != nil {
					// Pass the error up the error channel.
					errc <- err
					break
				}

				// Send process to the out channel.
				cs.Out <- cs
			case <-cs.context.Done():
				return
			}
		}

	}()

	return errc, nil
}

func (cs *Phpcs) process() error {

	return nil
}
