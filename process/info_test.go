package process

import (
	"context"
	"reflect"
	"testing"
	"time"
	"github.com/wptide/pkg/message"
	"github.com/wptide/pkg/tide"
	"bytes"
	"os"
	"github.com/wptide/pkg/log"
)

func TestInfo_Run(t *testing.T) {

	b := bytes.Buffer{}
	log.SetOutput(&b)
	defer log.SetOutput(os.Stdout)

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
						Result:    make(map[string]interface{}),
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
						Result:    make(map[string]interface{}),
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
						Result:    make(map[string]interface{}),
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
						Result:  make(map[string]interface{}),
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
						Result:    make(map[string]interface{}),
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
				err = info.Run(&errc)
			}()

			// Sleep a short time delay to give process time to start.
			time.Sleep(time.Millisecond * 100)

			if (err != nil) != tt.wantErr {
				t.Errorf("Info.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (chanError != nil) != tt.wantErrc {
				t.Errorf("Info.Run() errorChan = %v, wantErrc %v", chanError, tt.wantErrc)
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
