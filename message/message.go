package message

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
	Audits    *[]Audit `json:"audits,omitempty"`
}

// @todo: Legacy. Needs to be deprecated over time.
type Audit struct {
	Type    string       `json:"type"`
	Options *AuditOption `json:"options,omitempty"`
}

// @todo: Legacy. Needs to be deprecated over time.
type AuditOption struct {
	Standard   string `json:"standard"`
	Report     string `json:"report"`
	Encoding   string `json:"encoding,omitempty"`
	RuntimeSet string `json:"runtime-set,omitempty"`
	Ignore     string `json:"ignore,omitempty"`
}

type MessageProvider interface {
	SendMessage(msg *Message) error
	GetNextMessage() (*Message, error)
	DeleteMessage(ref *string) error
}
