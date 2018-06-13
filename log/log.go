package log

import (
	"io"
	"log"
	"os"
)

var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)

// Log writes to the logger with a title and content.
func Log(title interface{}, content interface{}) {
	logger.Printf("%s: %s", title, content)
}

// SetOutput changes the io.Writer of the logger.
func SetOutput(w io.Writer) {
	logger.SetOutput(w)
}
