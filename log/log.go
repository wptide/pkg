package log

import (
	"log"
	"os"
	"io"
)

var logger = log.New(os.Stdout,"", log.Ldate | log.Ltime )

func Log(title interface{}, content interface{}) {
	logger.Printf("%s: %s", title, content)
}

func SetOutput(w io.Writer) {
	logger.SetOutput(w)
}
