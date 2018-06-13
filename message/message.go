package message

// QueueMessage defines how messages are stored in a document store.
type QueueMessage struct {
	Created        int64    `json:"created" firestore:"created"`
	Lock           int64    `json:"lock" firestore:"lock"`
	Message        *Message `json:"message" firestore:"message"`
	Retries        int64    `json:"retries" firestore:"retries"`
	Status         string   `json:"status" firestore:"status"`
	RetryAvailable bool     `json:"retry_available" firestore:"retry_available"`
}

// Message represents a task to read from or send to a queue.
type Message struct {
	ResponseAPIEndpoint string  `json:"response_api_endpoint"`
	PayloadType         string  `json:"payload_type"`
	Title               string  `json:"title"`
	Content             string  `json:"content"`
	Slug                string  `json:"slug"`
	ProjectType         string  `json:"project_type,omitempty"`
	SourceURL           string  `json:"source_url"`
	SourceType          string  `json:"source_type"`
	RequestClient       string  `json:"request_client"`
	Force               bool    `json:"force"`
	Visibility          string  `json:"visibility"`
	ExternalRef         *string `json:"external_ref,omitempty"`
	// @todo: Legacy fields. Need to deprecate over time.
	Standards []string `json:"standards,omitempty"`
	Audits    []*Audit `json:"audits,omitempty"`
}

// Audit describes an audit type with its options.
type Audit struct {
	Type    string       `json:"type"`
	Options *AuditOption `json:"options,omitempty"`
}

// AuditOption describes specific options for an Audit.
type AuditOption struct {
	Standard         string `json:"standard,omitempty"`
	Report           string `json:"report,omitempty"`
	Encoding         string `json:"encoding,omitempty"`
	RuntimeSet       string `json:"runtime-set,omitempty"`
	Ignore           string `json:"ignore,omitempty"`
	StandardOverride string `json:"standard-override,omitempty"`
}

// Provider is an interface for creating new providers. E.g. firestore, mongo, sqs.
type Provider interface {
	SendMessage(msg *Message) error
	GetNextMessage() (*Message, error)
	DeleteMessage(ref *string) error
	Close() error
}
