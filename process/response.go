package process

import (
	"errors"
	"fmt"

	"github.com/wptide/pkg/payload"
)

// Response defines the structure for a Response process.
// This determines where the processed results will be sent.
type Response struct {
	Process                                 // Inherits methods from Process.
	In         <-chan Processor             // Expects a processor channel as input.
	Out        chan Processor               // (Optional) Send results to an output channel.
	Payloaders map[string]payload.Payloader // A map of "Payloader"s for different services.
}

// Run executes the process in a pipe.
func (res *Response) Run(errc *chan error) error {

	if res.In == nil {
		return errors.New("requires a previous process")
	}

	if len(res.Payloaders) == 0 {
		return errors.New("need to provide at least one payload manager")
	}

	go func() {
		for {
			select {
			case in := <-res.In:

				// Copy Process fields from `in` process.
				res.CopyFields(in)

				// Run the process.
				// If processing produces an error send it up the error channel.
				if err := res.Do(); err != nil {
					// Pass the error up the error channel.
					*errc <- errors.New("Response Error: " + err.Error())
					// Don't break, the message is still useful to other processes.
				}

				// Send process to the out channel.
				if res.Out != nil {
					res.Out <- res
				}
			}
		}

	}()

	return nil
}

// Do executes the process.
func (res *Response) Do() error {

	result := *res.Result

	payloadType := res.Message.PayloadType
	if payloadType == "" {
		// This is temporary, in future there will be no fallback.
		// Ensure all tasks include the PayloadType.
		// @todo An empty payloadType should return an error in future.
		payloadType = "tide"
	}

	payloader, ok := res.Payloaders[payloadType]
	if !ok {
		return errors.New("Could not find a valid payload generator for task")
	}

	p, err := payloader.BuildPayload(res.Message, result)
	if err != nil {
		return err
	}

	reply, err := payloader.SendPayload(res.Message.ResponseAPIEndpoint, p)
	if err != nil {
		return err
	}

	result["response"] = string(reply)
	result["responseMessage"] = fmt.Sprintf("'%s' payload submitted successfully.", payloadType)
	result["responseSuccess"] = true

	res.Result = &result

	return nil
}
