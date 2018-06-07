package pipe

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/wptide/pkg/message"
	"github.com/wptide/pkg/process"
)

type mockProcess struct {
	option    string
	shouldErr bool
}

func (m mockProcess) Run(errc *chan error) error {

	switch m.option {
	case "error":
		*errc <- errors.New("something went wrong")
	}

	if m.shouldErr {
		return errors.New("something went wrong with setup")
	}

	return nil
}

func (m mockProcess) SetContext(ctx context.Context) {}
func (m mockProcess) SetMessage(msg message.Message) {}
func (m mockProcess) GetMessage() message.Message    { return message.Message{} }
func (m mockProcess) SetResults(res *process.Result) {}
func (m mockProcess) GetResult() *process.Result     { return nil }
func (m mockProcess) SetFilesPath(path string)       {}
func (m mockProcess) GetFilesPath() string           { return "" }
func (m mockProcess) Do() error                      { return nil }

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want reflect.Type
	}{
		{
			"New Pipe",
			reflect.TypeOf(&Pipe{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(); !reflect.DeepEqual(reflect.TypeOf(got), tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithProcesses(t *testing.T) {

	proc := &mockProcess{}
	testPipe := &Pipe{
		processes: []process.Processor{
			proc,
		},
	}
	testPipe.init()

	type args struct {
		procs []process.Processor
	}
	tests := []struct {
		name string
		args args
		want *Pipe
	}{
		{
			"New pipe WithProcesses",
			args{
				[]process.Processor{
					proc,
				},
			},
			testPipe,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithProcesses(tt.args.procs...); !reflect.DeepEqual(got.processes, tt.want.processes) {
				t.Errorf("WithProcesses() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipe_AddProcesses(t *testing.T) {

	testPipe := &Pipe{}
	testPipe.init()

	type args struct {
		procs []process.Processor
	}
	tests := []struct {
		name    string
		p       *Pipe
		args    args
		wantErr bool
	}{
		{
			"Valid Processes",
			testPipe,
			args{
				[]process.Processor{
					&mockProcess{},
				},
			},
			false,
		},
		{
			"Invalid Processes",
			testPipe,
			args{
				[]process.Processor{
					nil,
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.AddProcesses(tt.args.procs...); (err != nil) != tt.wantErr {
				t.Errorf("Pipe.AddProcesses() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPipe_Run(t *testing.T) {

	tests := []struct {
		name     string
		p        *Pipe
		procs    []process.Processor
		wantErr  bool
		wantErrc bool
	}{
		{
			"Run with valid process",
			&Pipe{},
			[]process.Processor{
				&mockProcess{},
			},
			false,
			false,
		},
		{
			"Run with invalid process",
			&Pipe{},
			[]process.Processor{
				&mockProcess{
					option: "error",
				},
			},
			false,
			true,
		},
		{
			"Run with invalid process setup",
			&Pipe{},
			[]process.Processor{
				&mockProcess{
					shouldErr: true,
				},
			},
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.p.init()
			tt.p.AddProcesses(tt.procs...)

			var err error
			var chanError error
			errc := make(chan error)

			go func() {
				for {
					select {
					case e := <-errc:
						chanError = e
					}
				}
			}()

			go func() {
				err = tt.p.Run(&errc)
			}()

			// Sleep a short time delay to give process time to start.
			time.Sleep(time.Millisecond * 100)

			if (err != nil) != tt.wantErr {
				t.Errorf("Pipe.Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if (chanError != nil) != tt.wantErrc {
				t.Errorf("Pipe.Run() errorChan = %v, wantErrc %v", chanError, tt.wantErrc)
			}
		})
	}
}
