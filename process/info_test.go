package process

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/wptide/pkg/audit"
	"github.com/wptide/pkg/message"
	"github.com/wptide/pkg/tide"
)

func TestInfo_Run(t *testing.T) {

	// Need to test with a context.
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	type fields struct {
		Process Process
		In      <-chan Processor
		Out     chan Processor
	}
	tests := []struct {
		name     string
		fields   fields
		procs    []Processor
		wantErrc bool
		wantErr  bool
	}{
		{
			"Invalid In channel",
			fields{
				Out: make(chan Processor),
			},
			nil,
			false,
			true,
		},
		{
			"Invalid Out channel",
			fields{
				In: make(chan Processor),
			},
			nil,
			false,
			true,
		},
		{
			"Plugin",
			fields{
				In:  make(<-chan Processor),
				Out: make(chan Processor),
			},
			[]Processor{
				&Ingest{
					Process: Process{
						Message:   message.Message{Title: "Test Plugin"},
						FilesPath: "./testdata/info/plugin",
						Result:    audit.Result{},
					},
				},
			},
			false,
			false,
		},
		{
			"Theme",
			fields{
				In:  make(<-chan Processor),
				Out: make(chan Processor),
			},
			[]Processor{
				&Ingest{
					Process: Process{
						Message:   message.Message{Title: "Test Theme"},
						FilesPath: "./testdata/info/theme",
						Result:    audit.Result{},
					},
				},
			},
			false,
			false,
		},
		{
			"Other",
			fields{
				In:  make(<-chan Processor),
				Out: make(chan Processor),
			},
			[]Processor{
				&Ingest{
					Process: Process{
						Message:   message.Message{Title: "Test Other"},
						FilesPath: "./testdata/info/other",
						Result:    audit.Result{},
					},
				},
			},
			false,
			false,
		},
		{
			"No Files Path",
			fields{
				In:  make(<-chan Processor),
				Out: make(chan Processor),
			},
			[]Processor{
				&Ingest{
					Process: Process{
						Message: message.Message{Title: "No Files Path"},
						Result:  audit.Result{},
					},
				},
			},
			true,
			false,
		},
		{
			"Invalid Path",
			fields{
				In:  make(<-chan Processor),
				Out: make(chan Processor),
			},
			[]Processor{
				&Ingest{
					Process: Process{
						Message:   message.Message{Title: "Invalid Path"},
						FilesPath: "./testdata/info/invalid",
						Result:    audit.Result{},
					},
				},
			},
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			info := &Info{
				Process: tt.fields.Process,
				In:      tt.fields.In,
				Out:     tt.fields.Out,
			}

			info.SetContext(ctx)
			if tt.procs != nil && len(tt.procs) != 0 {
				info.In = generateProcs(ctx, tt.procs)
			}

			var errc <-chan error
			var err error

			go func() {
				errc, err = info.Run()
			}()

			// Sleep a short time delay to give process time to start.
			time.Sleep(time.Millisecond * 100)

			if (err != nil) != tt.wantErr {
				t.Errorf("Info.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (len(errc) != 0) != tt.wantErrc {
				e := <-errc
				t.Errorf("Info.Run() error = %v, wantErrc %v", e, tt.wantErrc)
				return
			}

		})
	}
}

func Test_getProjectDetails(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   []tide.InfoDetails
		wantErr bool
	}{
		{
			"Invalid directory",
			args{
				"./testdata/invalid",
			},
			"",
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getProjectDetails(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("getProjectDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getProjectDetails() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("getProjectDetails() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
