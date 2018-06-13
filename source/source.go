package source

import (
	"strings"
)

// Source interface describes the source for code to be audited.
type Source interface {
	PrepareFiles(dest string) error
	GetChecksum() string
	GetFiles() []string
}

// GetKind uses basic string manipulation to get the type of source file.
func GetKind(url string) string {
	var kind string
	ts := strings.Split(url, ".")
	if len(ts) > 1 {
		kind = ts[len(ts)-1]
	}
	return kind
}
