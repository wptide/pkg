package process

import (
	"testing"
	"context"
	"github.com/wptide/pkg/message"
	"time"
)

func TestTide_Run(t *testing.T) {

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
				Out:        make(chan Processor),
			},
			nil,
			false,
			true,
		},
		{
			"Invalid Out channel",
			fields{
				In:         make(chan Processor),
			},
			nil,
			false,
			true,
		},
		{
			"Test",
			fields{
				In:  make(<-chan Processor),
				Out: make(chan Processor),
			},
			[]Processor{
				&Ingest{
					Process: Process{ Message:message.Message{Title:"Test"} },
				},
			},
			false,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &Tide{
				Process: tt.fields.Process,
				In:      tt.fields.In,
				Out:     tt.fields.Out,
			}

			tc.SetContext(ctx)
			if tt.procs != nil {
				tc.In = generateProcs(ctx,tt.procs)
			}

			var errc <-chan error
			var err error

			go func() {
				errc, err = tc.Run()
			}()

			// Sleep a short time delay to give process time to start.
			time.Sleep(time.Millisecond * 100)

			if (err != nil) != tt.wantErr {
				t.Errorf("Phpcs.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}