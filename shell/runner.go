package shell

import (
	"os/exec"
	"io/ioutil"
	"sync"
	"io"
	"errors"
	"syscall"
)

type Runner interface {
	Run(name string, arg ...string) ([]byte, []byte, error, int)
}

type Command struct {
	execFunc   func(name string, arg ...string) *exec.Cmd
	Cmd        *exec.Cmd
	once       sync.Once
	stdOutFunc func(*exec.Cmd) (io.ReadCloser, error)
	stdErrFunc func(*exec.Cmd) (io.ReadCloser, error)
	startFunc  func(*exec.Cmd) (error)
	waitFunc   func(*exec.Cmd) (error)
}

func (c *Command) Run(name string, arg ...string) ([]byte, []byte, error, int) {

	// Init execFunc if not set. Once.
	c.once.Do(func() {
		c.PrepareFuncs()
	})

	// Prepare the command and set the stdOut pipe.
	cmd := c.execFunc(name, arg...)

	cmdReader, err := c.stdOutFunc(cmd)
	if err != nil {
		return nil, nil, err, 2
	}

	errReader, err := c.stdErrFunc(cmd)
	if err != nil {
		return nil, nil, err, 2
	}

	// Start the command.
	err = c.startFunc(cmd)
	if err != nil {
		return nil, nil, err, 2
	}

	// Read stdout pipe.
	// We already checked for a valid pipe, so skipping error check.
	resultBytes, _ := ioutil.ReadAll(cmdReader)

	// Read stderr pipe.
	// We already checked for a valid pipe, so skipping error check.
	errorBytes, _ := ioutil.ReadAll(errReader)

	// Wait for command to exit and stdio to be read.
	exitCode := 0
	exitErr := c.waitFunc(cmd)
	if exitErr, ok := exitErr.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			exitCode = status.ExitStatus()
		}
		return resultBytes, errorBytes, exitErr, exitCode
	}

	return resultBytes, errorBytes, nil, exitCode
}

func (c *Command) PrepareFuncs() {
	if c.execFunc == nil {
		c.SetExecFunc(exec.Command)
	}
	if c.stdOutFunc == nil {
		c.stdOutFunc = func(cmd *exec.Cmd) (io.ReadCloser, error) {
			if cmd == nil {
				return nil, errors.New("command not provided")
			}
			return cmd.StdoutPipe()
		}
	}
	if c.stdErrFunc == nil {
		c.stdErrFunc = func(cmd *exec.Cmd) (io.ReadCloser, error) {
			if cmd == nil {
				return nil, errors.New("command not provided")
			}
			return cmd.StderrPipe()
		}
	}
	if c.startFunc == nil {
		c.startFunc = func(cmd *exec.Cmd) error {
			if cmd == nil {
				return errors.New("command not provided")
			}
			return cmd.Start()
		}
	}
	if c.waitFunc == nil {
		c.waitFunc = func(cmd *exec.Cmd) error {
			if cmd == nil {
				return errors.New("command not provided")
			}
			return cmd.Wait()
		}
	}
}

func (c *Command) SetExecFunc(execFunc func(name string, arg ...string) *exec.Cmd) {
	c.execFunc = execFunc
}
