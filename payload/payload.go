package payload

import (
	"github.com/wptide/pkg/message"
)

// Sender interface describes a send message to an endpoint.
type Sender interface {
	SendPayload(destination string, payload []byte) ([]byte, error)
}

// Builder interface describes a payload generator.
type Builder interface {
	BuildPayload(message.Message, map[string]interface{}) ([]byte, error)
}

// Payloader interface is used to build and send payloads to endpoints.
type Payloader interface {
	Sender
	Builder
}
