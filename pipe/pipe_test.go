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

func (m mockProcess) Run() (<-chan error, error) {
	errc := make(chan error, 1)

	switch m.option {
	case "error":
		errc <- errors.New("something went wrong")
	}

	if m.shouldErr {
		return errc, errors.New("something went wrong with setup")
	}

	return errc, nil
}

func (m mockProcess) SetContext(ctx context.Context)        {}
func (m mockProcess) SetMessage(msg message.Message)        {}
func (m mockProcess) GetMessage() message.Message           { return message.Message{} }
func (m mockProcess) SetResults(res map[string]interface{}) {}
func (m mockProcess) GetResult() map[string]interface{}     { return nil }
func (m mockProcess) SetFilesPath(path string)              {}
func (m mockProcess) GetFilesPath() string                  { return "" }

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
			"Valid Processes",
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
		name    string
		p       *Pipe
		procs   []process.Processor
		wantErr bool
	}{
		{
			"Run with valid process",
			&Pipe{},
			[]process.Processor{
				&mockProcess{},
			},
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.p.init()
			tt.p.AddProcesses(tt.procs...)

			var err error

			go func() {
				err = tt.p.Run()
			}()

			// Sleep a short time delay to give process time to start.
			time.Sleep(time.Millisecond * 100)

			if (err != nil) != tt.wantErr {
				t.Errorf("Pipe.Run() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}

func TestPipe_wait(t *testing.T) {
	tests := []struct {
		name    string
		p       Pipe
		wantErr bool
	}{
		{
			"Execute wait()",
			Pipe{},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.wait(); (err != nil) != tt.wantErr {
				t.Errorf("Pipe.wait() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPipe_mergeErrors(t *testing.T) {

	errc1 := make(chan error, 1)
	errc2 := make(chan error, 1)

	tests := []struct {
		name    string
		p       Pipe
		want    reflect.Type
		hasErr1 bool
		hasErr2 bool
	}{
		{
			"No error chans",
			Pipe{},
			reflect.TypeOf(make(<-chan error)),
			false,
			false,
		},
		{
			"Empty error chans",
			Pipe{
				errors: []<-chan error{},
			},
			reflect.TypeOf(make(<-chan error)),
			false,
			false,
		},
		{
			"No items - error chans",
			Pipe{
				errors: []<-chan error{
					make(<-chan error),
					make(<-chan error),
				},
			},
			reflect.TypeOf(make(<-chan error)),
			false,
			false,
		},
		{
			"With items - both chans",
			Pipe{
				errors: []<-chan error{
					errc1,
					errc2,
				},
			},
			reflect.TypeOf(make(<-chan error)),
			true,
			true,
		},
		{
			"With items - one chan",
			Pipe{
				errors: []<-chan error{
					errc1,
					errc2,
				},
			},
			reflect.TypeOf(make(<-chan error)),
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.hasErr1 {
				errc1 <- errors.New("Error chan 1")
			}

			if tt.hasErr2 {
				errc2 <- errors.New("Error chan 2")
			}

			if got := tt.p.mergeErrors(); !reflect.DeepEqual(reflect.TypeOf(got), tt.want) {
				t.Errorf("Pipe.mergeErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}
