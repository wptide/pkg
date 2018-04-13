package process

import (
	"errors"
	"github.com/wptide/pkg/payload"
	"fmt"
)

// Response defines the structure for a Response process.
// This determines where the processed results will be sent.
type Response struct {
	Process                                 // Inherits methods from Process.
	In         <-chan Processor             // Expects a processor channel as input.
	Out        chan Processor               // (Optional) Send results to an output channel.
	Payloaders map[string]payload.Payloader // A map of "Payloader"s for different services.
}

func (res *Response) Run() (<-chan error, error) {

	if res.In == nil {
		return nil, errors.New("requires a previous process")
	}

	if len(res.Payloaders) == 0 {
		return nil, errors.New("need to provide at least one payload manager")
	}

	errc := make(chan error, 1)

	go func() {
		defer close(errc)
		for {
			select {
			case in := <-res.In:

				// Copy Process fields from `in` process.
				res.CopyFields(in)

				// Run the process.
				// If processing produces an error send it up the error channel.
				if err := res.process(); err != nil {
					// Pass the error up the error channel.
					errc <- err
					// Don't break, the message is still useful to other processes.
				}

				// Send process to the out channel.
				if res.Out != nil {
					res.Out <- res
				}
			}
		}

	}()

	return errc, nil
}

func (res *Response) process() error {

	payloadType := res.Message.PayloadType
	if payloadType == "" {
		// This is temporary, in future there will be no fallback.
		// Ensure all tasks include the PayloadType.
		// @todo An empty payloadType should return an error in future.
		payloadType = "tide"
	}

	payloader, ok := res.Payloaders[payloadType]
	if ! ok {
		return errors.New("Could not find a valid payload generator for task")
	}

	payload, err := payloader.BuildPayload(res.Message, res.Result)
	if err != nil {
		return err
	}

	reply, err := payloader.SendPayload(res.Message.ResponseAPIEndpoint, payload)
	if err != nil {
		return err
	}

	res.Result["response"] = string(reply)
	res.Result["responseMessage"] = fmt.Sprintf("'%s' payload submitted successfully.", payloadType)
	res.Result["responseSuccess"] = true

	return nil
}
