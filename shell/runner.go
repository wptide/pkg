package shell

import (
	"os/exec"
	"sync"
	"syscall"
	"bytes"
)

type Runner interface {
	Run(name string, arg ...string) ([]byte, []byte, error, int)
}

type Command struct {
	execFunc   func(name string, arg ...string) *exec.Cmd
	once       sync.Once
}

func (c *Command) Run(name string, arg ...string) ([]byte, []byte, error, int) {

	c.once.Do(func() {
		if c.execFunc == nil {
			c.execFunc = exec.Command
		}
	})

	resultsBuffer := bytes.Buffer{}
	errorsBuffer := bytes.Buffer{}
	cmd:= c.execFunc(name, arg...)
	cmd.Stdout = &resultsBuffer
	cmd.Stderr = &errorsBuffer

	exitCode := 0
	exitErr := cmd.Run()

	if exitErr, ok := exitErr.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			exitCode = status.ExitStatus()
		}
	}

	return resultsBuffer.Bytes(), errorsBuffer.Bytes(), exitErr, exitCode
}