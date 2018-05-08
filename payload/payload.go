package payload

import (
	"github.com/wptide/pkg/message"
)

type PayloadSender interface {
	SendPayload(destination string, payload []byte) ([]byte, error)
}

type PayloadBuilder interface {
	BuildPayload(message.Message, map[string]interface{}) ([]byte, error)
}

type Payloader interface {
	PayloadSender
	PayloadBuilder
}
