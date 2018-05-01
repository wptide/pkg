package process

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/wptide/pkg/message"
	"github.com/wptide/pkg/source"
	"context"
	"time"
	"github.com/wptide/pkg/log"
	"bytes"
)

type mockSource struct{}

func (m mockSource) PrepareFiles(dest string) error { return nil }
func (m mockSource) GetChecksum() string            { return "" }
func (m mockSource) GetFiles() []string             { return nil }

type mockProcess struct {
	Process
	In  <-chan Processor
	Out chan Processor
}

func (m *mockProcess) Run() (<-chan error, error)     { return nil, nil }
func (m *mockProcess) SetContext(ctx context.Context) {}

var ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	switch r.URL.String() {
	case "/test.zip":
		http.ServeFile(w, r, "./testdata/test.zip")
		return
	case "/api/audits":
		http.ServeFile(w, r, `{ "message": "Payload received" }`)
		return
	default:
		fmt.Fprintln(w, "Nothing to see here.")
		return
	}
}))

func Test_validateMessage(t *testing.T) {

	b := bytes.Buffer{}
	log.SetOutput(&b)
	defer log.SetOutput(os.Stdout)

	type args struct {
		msg message.Message
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"Valid message",
			args{
				message.Message{
					Title:               "Valid Message",
					ResponseAPIEndpoint: "http://test.local",
					SourceURL:           "http://test.local/source.zip",
					SourceType:          "zip",
				},
			},
			false,
		},
		{
			"Missing Title",
			args{
				message.Message{},
			},
			true,
		},
		{
			"Missing Response Endpoint",
			args{
				message.Message{
					Title: "Valid Title",
				},
			},
			true,
		},
		{
			"Missing Source URL",
			args{
				message.Message{
					Title:               "Valid Title",
					ResponseAPIEndpoint: "http://test.local",
				},
			},
			true,
		},
		{
			"Missing Source Type",
			args{
				message.Message{
					Title:               "Valid Title",
					ResponseAPIEndpoint: "http://test.local",
					SourceURL:           "http://test.local/source.zip",
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateMessage(tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("validateMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIngest_process(t *testing.T) {

	b := bytes.Buffer{}
	log.SetOutput(&b)
	defer log.SetOutput(os.Stdout)

	// Make a /tmp folder
	os.Mkdir("./testdata/tmp", os.ModePerm)

	// Clean up after.
	defer func() {
		os.RemoveAll("./testdata/tmp")
	}()

	type options struct {
		tempFolder string
		sourceMgr  source.Source
	}

	tests := []struct {
		name    string
		message message.Message
		options options
		wantErr bool
	}{
		{
			"Valid Ingest",
			message.Message{
				Title:               "Test Ingest",
				ResponseAPIEndpoint: ts.URL + "/api/audits",
				SourceURL:           ts.URL + "/test.zip",
				SourceType:          "zip",
			},
			options{},
			false,
		},
		{
			"No valid source manager",
			message.Message{
				Title:               "Invalid Source Manager",
				ResponseAPIEndpoint: ts.URL + "/api/audits",
				SourceURL:           ts.URL + "/test.rar",
				SourceType:          "rar",
			},
			options{},
			true,
		},
		{
			"Source not valid",
			message.Message{
				Title:               "Invalid Source",
				ResponseAPIEndpoint: ts.URL + "/api/audits",
				SourceURL:           ts.URL + "/notfound.zip",
				SourceType:          "zip",
			},
			options{},
			true,
		},
		{
			"Empty Checksum",
			message.Message{
				Title:               "Invalid Checksum",
				ResponseAPIEndpoint: ts.URL + "/api/audits",
				SourceURL:           ts.URL + "/empty.fake", // Use a fake "extension".
				SourceType:          "fake",
			},
			options{
				sourceMgr: mockSource{},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ig := &Ingest{
				TempFolder: "./testdata/tmp",
			}
			ig.Result = make(map[string]interface{})

			if tt.options.tempFolder != "" {
				ig.TempFolder = tt.options.tempFolder
			}

			if tt.options.sourceMgr != nil {
				ig.sourceManager = tt.options.sourceMgr
			}

			ig.Message = tt.message
			if err := ig.process(); (err != nil) != tt.wantErr {
				t.Errorf("Ingest.process() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIngest_Run(t *testing.T) {

	b := bytes.Buffer{}
	log.SetOutput(&b)
	defer log.SetOutput(os.Stdout)

	// Make a /tmp folder
	os.Mkdir("./testdata/tmp", os.ModePerm)

	// Clean up after.
	defer func() {
		os.RemoveAll("./testdata/tmp")
	}()

	// Need to test with a context.
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	type fields struct {
		Process    Process
		In         <-chan message.Message
		Out        chan Processor
		TempFolder string
		srcMgr     source.Source
	}

	tests := []struct {
		name     string
		fields   fields
		messages []message.Message
		wantErrc bool
		wantErr  bool
	}{
		{
			"Invalid In channel",
			fields{
				Out:        make(chan Processor),
				TempFolder: "./testdata/tmp",
			},
			nil,
			false,
			true,
		},
		{
			"Invalid Out channel",
			fields{
				In:         make(chan message.Message),
				TempFolder: "./testdata/tmp",
			},
			nil,
			false,
			true,
		},
		{
			"No TempFolder",
			fields{
				In:  make(chan message.Message),
				Out: make(chan Processor),
			},
			nil,
			false,
			true,
		},
		{
			"Valid Messages",
			fields{
				In:         make(chan message.Message),
				Out:        make(chan Processor),
				TempFolder: "./testdata/tmp",
			},
			[]message.Message{
				{
					Title:               "Message",
					ResponseAPIEndpoint: "http://test.local/api/audits/",
					SourceURL:           ts.URL + "/test.zip",
					SourceType:          "zip",
				},
			},
			false,
			false,
		},
		{
			// Don't provide any messages. This should cause
			// the context to close.
			"No Messages",
			fields{
				In:         make(chan message.Message),
				Out:        make(chan Processor),
				TempFolder: "./testdata/tmp",
			},
			nil,
			false,
			false,
		},
		{
			"Invalid Messages",
			fields{
				In:         make(chan message.Message),
				Out:        make(chan Processor),
				TempFolder: "./testdata/tmp",
			},
			[]message.Message{
				{
					ResponseAPIEndpoint: "http://test.local/api/audits/",
					SourceURL:           ts.URL + "/test.zip",
					SourceType:          "zip",
				},
			},
			true,
			false,
		},
		{
			"Invalid Source Message",
			fields{
				In:         make(chan message.Message),
				Out:        make(chan Processor),
				TempFolder: "./testdata/tmp",
			},
			[]message.Message{
				{
					Title: "Invalid Source",
					ResponseAPIEndpoint: "http://test.local/api/audits/",
					SourceURL:           ts.URL + "/test.rar",
					SourceType:          "rar",
				},
			},
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ig := &Ingest{
				Process:       tt.fields.Process,
				In:            tt.fields.In,
				Out:           tt.fields.Out,
				TempFolder:    tt.fields.TempFolder,
				sourceManager: tt.fields.srcMgr,
			}

			ig.SetContext(ctx)
			if tt.messages != nil {
				ig.In = generateMessages(tt.messages)
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
				err = ig.Run(&errc)
			}()

			// Sleep a short time delay to give process time to start.
			time.Sleep(time.Millisecond * 100)

			if (err != nil) != tt.wantErr {
				t.Errorf("Ingest.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If there is an item in the output channel read it to satisfy coverage.
			if len(ig.Out) != 0 {
				<-ig.Out
			}

			if (chanError != nil) != tt.wantErrc {
				t.Errorf("Info.Run() errorChan = %v, wantErrc %v", chanError, tt.wantErrc)
			}
		})
	}
}

func generateMessages(messages []message.Message) <-chan message.Message {
	out := make(chan message.Message, len(messages))

	go func() {
		defer close(out)
		for _, msg := range messages {
			out <- msg
		}
	}()

	return out
}
