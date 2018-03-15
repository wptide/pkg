package shell

import (
	"os/exec"
	"testing"
	"os"
	"fmt"
	"io"
	"github.com/pkg/errors"
)

func mockExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestCommand_Run(t *testing.T) {
	type args struct {
		name string
		arg  []string
	}

	type funcs struct {
		stdOutFunc func(*exec.Cmd) (io.ReadCloser, error)
		stdErrFunc func(*exec.Cmd) (io.ReadCloser, error)
		startFunc  func(*exec.Cmd) (error)
		waitFunc   func(*exec.Cmd) (error)
	}

	tests := []struct {
		name     string
		c        *Command
		funcs    funcs
		args     args
		wantOutB []byte
		wantErrB []byte
		wantErr  bool
	}{
		{
			"Run Success",
			&Command{
				execFunc: mockExecCommand,
			},
			funcs{},
			args{
				name: "test-success",
			},
			[]byte("Success!"),
			nil,
			false,
		},
		{
			"Run Fail",
			&Command{
				execFunc: mockExecCommand,
			},
			funcs{},
			args{
				name: "test-fail",
			},
			nil,
			[]byte("Failed!"),
			false,
		},
		{
			"StdOut fail",
			&Command{
				execFunc: mockExecCommand,
			},
			funcs{
				stdOutFunc: func(cmd *exec.Cmd) (io.ReadCloser, error) {
					return nil, errors.New("Stdout Pipe Failed.")
				},
			},
			args{
				name: "fail",
			},
			nil,
			nil,
			true,
		},
		{
			"StdErr fail",
			&Command{
				execFunc: mockExecCommand,
			},
			funcs{
				stdErrFunc: func(cmd *exec.Cmd) (io.ReadCloser, error) {
					return nil, errors.New("Stderr Pipe Failed.")
				},
			},
			args{
				name: "fail",
			},
			nil,
			nil,
			true,
		},
		{
			"Start fail",
			&Command{
				execFunc: mockExecCommand,
			},
			funcs{
				startFunc: func(cmd *exec.Cmd) error {
					return errors.New("Start failed")
				},
			},
			args{
				name: "fail",
			},
			nil,
			nil,
			true,
		},
		{
			"Wait fail",
			&Command{
				execFunc: mockExecCommand,
			},
			funcs{
				waitFunc: func(cmd *exec.Cmd) error {
					return errors.New("Wait failed")
				},
			},
			args{
				name: "fail",
			},
			nil,
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.funcs.stdOutFunc != nil {
				tt.c.stdOutFunc = tt.funcs.stdOutFunc
			}

			if tt.funcs.stdErrFunc != nil {
				tt.c.stdErrFunc = tt.funcs.stdErrFunc
			}

			if tt.funcs.startFunc != nil {
				tt.c.startFunc = tt.funcs.startFunc
			}

			if tt.funcs.waitFunc != nil {
				tt.c.waitFunc = tt.funcs.waitFunc
			}

			outBuff, errBuff, err := tt.c.Run(tt.args.name, tt.args.arg...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Command.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(outBuff) != string(tt.wantOutB) {
				t.Errorf("Command.Run() outBuff = %v, want %v", string(outBuff), string(tt.wantOutB))
			}
			if string(errBuff) != string(tt.wantErrB) {
				t.Errorf("Command.Run() errBuff = %v, want %v", string(errBuff), string(tt.wantErrB))
			}
		})
	}
}

func TestCommand_PrepareFuncs(t *testing.T) {
	tests := []struct {
		name    string
		c       *Command
		command string
		wantErr bool
	}{
		{
			"Commands fail",
			&Command{},
			"",
			true,
		},
		{
			"Commands success",
			&Command{},
			"test-success",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.c.PrepareFuncs()

			var cmd *exec.Cmd = nil
			if tt.command != "" {
				cmd = mockExecCommand(tt.command)
			}

			if _, err := tt.c.stdOutFunc(cmd); (err != nil) != tt.wantErr {
				t.Errorf("Command.stdOutFunc() err: %v", err)
			}

			if _, err := tt.c.stdErrFunc(cmd); (err != nil) != tt.wantErr {
				t.Errorf("Command.stdErrFunc() err: %v", err)
			}

			if err := tt.c.startFunc(cmd); (err != nil) != tt.wantErr {
				t.Errorf("Command.startFunc() err: %v", err)
			}

			if err := tt.c.waitFunc(cmd); (err != nil) != tt.wantErr {
				t.Errorf("Command.waitFunc() err: %v", err)
			}
		})
	}
}

func TestCommand_SetExecFunc(t *testing.T) {
	type args struct {
		execFunc func(name string, arg ...string) *exec.Cmd
	}
	tests := []struct {
		name    string
		c       *Command
		args    args
		wantErr bool
	}{
		{
			"Nil func",
			&Command{},
			args{},
			true,
		},
		{
			"Valid func",
			&Command{},
			args{
				exec.Command,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.SetExecFunc(tt.args.execFunc)
			if (tt.c.execFunc == nil) != tt.wantErr {
				t.Errorf("Command.SetExecFunc() expects a valid %v", "func(name string, arg ...string) *exec.Cmd")
			}

		})
	}
}

// TestHelperProcess is the fake command.
func TestHelperProcess(t *testing.T) {
	// If the helper process var is not set this code should not run.
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// Exit helper sub routine if nothing else exits.
	defer os.Exit(0)

	// Get the passed arguments.
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}

	// If no arguments, write to Stderr and exit
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]

	switch cmd {

	case "test-success":
		fmt.Fprintf(os.Stdout, "Success!")
		os.Exit(0)
	case "test-fail":
		fmt.Fprintf(os.Stderr, "Failed!")
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %q\n", cmd)
		os.Exit(2)
	}
}