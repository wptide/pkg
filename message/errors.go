package message

// ProviderError is a new error type for message providers.
type ProviderError struct {
	error string
	Type  int
}

/*
 * Constants to represent error types.
 *
 * ErrCritical is a critical provider error.
 * ErrOverQuota is an over quota warning.
 */
const (
	ErrCritcal = iota
	ErrOverQuota
)

func (p ProviderError) Error() string {
	return p.error
}

// NewProviderError creates a new error object with the provided string as the message.
func NewProviderError(s string) *ProviderError {
	return &ProviderError{
		error: s,
	}
}
