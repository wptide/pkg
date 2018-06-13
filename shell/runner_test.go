package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"
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
		startFunc  func(*exec.Cmd) error
		waitFunc   func(*exec.Cmd) error
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
			"Default Exec Func - No Command",
			&Command{},
			funcs{},
			args{},
			nil,
			nil,
			true,
		},
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
			"Wait fail",
			&Command{
				execFunc: mockExecCommand,
			},
			funcs{},
			args{
				name: "test-exit",
			},
			[]byte("Exit!"),
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			outBuff, errBuff, _, err := tt.c.Run(tt.args.name, tt.args.arg...)
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
	case "test-exit":
		fmt.Fprintf(os.Stdout, "Exit!")
		os.Exit(22)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %q\n", cmd)
		os.Exit(2)
	}
}
