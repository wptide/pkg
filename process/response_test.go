package process

import (
	"testing"

	"github.com/wptide/pkg/payload"
	"context"
	"time"
	"github.com/wptide/pkg/message"
	"errors"
	"bytes"
	"os"
	"github.com/wptide/pkg/log"
)

type MockPayloader struct{}

func (m MockPayloader) BuildPayload(msg message.Message, data map[string]interface {}) ([]byte, error) {

	if msg.Slug == "buildFail" {
		return nil, errors.New("something went wrong")
	}

	payload := `{ "valid":"payload" }`
	return []byte(payload), nil
}

func (m MockPayloader) SendPayload(destination string, payload []byte) ([]byte, error) {

	if destination == "http://test.local/sendfail" {
		return nil, errors.New("something went wrong")
	}

	reply := `{ "status": "ok" }`
	return []byte(reply), nil
}

func TestResponse_Run(t *testing.T) {

	b := bytes.Buffer{}
	log.SetOutput(&b)
	defer log.SetOutput(os.Stdout)

	// Need to test with a context.
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	defaultPayloaders := map[string]payload.Payloader{
		"mock": MockPayloader{},
	}

	type fields struct {
		Process
		In         <-chan Processor
		Out        chan Processor
		Payloaders map[string]payload.Payloader
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
				Payloaders: defaultPayloaders,
			},
			nil,
			false,
			true,
		},
		{
			"Invalid Payloaders",
			fields{
				In: make(chan Processor),
			},
			nil,
			false,
			true,
		},
		{
			"Valid No Out Channel",
			fields{
				In:         make(<-chan Processor),
				Payloaders: defaultPayloaders,
			},
			[]Processor{
				&Ingest{
					Process: Process{
						Message: message.Message{
							Title:       "Test",
							PayloadType: "mock",
						},
						Result: &Result{},
					},
				},
			},
			false,
			false,
		},
		{
			"Invalid Payloader",
			fields{
				In:         make(<-chan Processor),
				Out:        make(chan Processor),
				Payloaders: defaultPayloaders,
			},
			[]Processor{
				&Ingest{
					Process: Process{
						Message: message.Message{
							Title:       "Test",
							PayloadType: "unknown",
						},
						Result: &Result{},
					},
				},
			},
			true,
			false,
		},
		{
			"Invalid Empty Payload Type",
			fields{
				In:         make(<-chan Processor),
				Out:        make(chan Processor),
				Payloaders: defaultPayloaders,
			},
			[]Processor{
				&Ingest{
					Process: Process{
						Message: message.Message{
							Title: "Test",
						},
						Result: &Result{},
					},
				},
			},
			true,
			false,
		},
		{
			"Payload Build Fail",
			fields{
				In:         make(<-chan Processor),
				Out:        make(chan Processor),
				Payloaders: defaultPayloaders,
			},
			[]Processor{
				&Ingest{
					Process: Process{
						Message: message.Message{
							Title:       "Payload Build Fail",
							Slug:        "buildFail",
							PayloadType: "mock",
						},
						Result: &Result{},
					},
				},
			},
			true,
			false,
		},
		{
			"Payload Send Fail",
			fields{
				In:         make(<-chan Processor),
				Out:        make(chan Processor),
				Payloaders: defaultPayloaders,
			},
			[]Processor{
				&Ingest{
					Process: Process{
						Message: message.Message{
							Title:       "Payload Send Fail",
							PayloadType: "mock",
							ResponseAPIEndpoint: "http://test.local/sendfail",
						},
						Result: &Result{},
					},
				},
			},
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &Response{
				Process:    tt.fields.Process,
				In:         tt.fields.In,
				Out:        tt.fields.Out,
				Payloaders: tt.fields.Payloaders,
			}

			tc.SetContext(ctx)
			if tt.procs != nil {
				tc.In = generateProcs(ctx, tt.procs)
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
				err = tc.Run(&errc)
			}()

			// Sleep a short time delay to give process time to start.
			time.Sleep(time.Millisecond * 100)

			if (err != nil) != tt.wantErr {
				t.Errorf("Response.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (chanError != nil) != tt.wantErrc {
				t.Errorf("Response.Run() errorChan = %v, wantErrc %v", chanError, tt.wantErrc)
			}
		})
	}
}