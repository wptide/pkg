package process

import (
	"testing"
	"context"
	"time"
	"github.com/wptide/pkg/storage"
	"os"
	"errors"
	"fmt"
	"io/ioutil"
	"github.com/wptide/pkg/shell"
	"github.com/wptide/pkg/message"
	"github.com/wptide/pkg/audit"
)

type mockRunner struct{}

func (m mockRunner) Run(name string, arg ...string) ([]byte, []byte, error, int) {

	switch arg[0] {
	case "https://wp-themes.com/test":
		return []byte(exampleReport()), nil, nil, 0
	case "https://wp-themes.com/jsonError":
		return []byte("this is not json"), nil, nil, 0
	case "https://wp-themes.com/error":
		return nil, []byte("error output"), nil, 0
	default:
		return nil, nil, errors.New("Something went wrong."), 1
	}
}

type mockStorage struct{}

func (m mockStorage) Kind() string {
	return "mock"
}

func (m mockStorage) CollectionRef() string {
	return "mock-collection"
}

func (m mockStorage) UploadFile(filename, reference string) error {
	return nil
}

func (m mockStorage) DownloadFile(reference, filename string) error {
	return nil
}

func mockWriteFile(filename string, data []byte, perm os.FileMode) error {
	switch filename {
	case "./testdata/tmp/ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff-lighthouse-full.json":
		return errors.New("something went wrong")
	default:
		return nil
	}
}

func TestLighthouse_Run(t *testing.T) {

	// Set out execCommand variable to the mock function.
	runner = &mockRunner{}
	// Remember to set it back after the test.
	defer func() { runner = &shell.Command{} }()

	// Set out execCommand variable to the mock function.
	writeFile = mockWriteFile
	// Remember to set it back after the test.
	defer func() { writeFile = ioutil.WriteFile }()

	// Need to test with a context.
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	type fields struct {
		Process         Process
		In              <-chan Processor
		Out             chan Processor
		TempFolder      string
		StorageProvider storage.StorageProvider
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
				Out:             make(chan Processor),
				StorageProvider: &mockStorage{},
				TempFolder:      "./testdata/tmp",
			},
			nil,
			false,
			true,
		},
		{
			"Invalid Out channel",
			fields{
				In:              make(chan Processor),
				StorageProvider: &mockStorage{},
				TempFolder:      "./testdata/tmp",
			},
			nil,
			false,
			true,
		},
		{
			"No Temp Folder",
			fields{
				In:              make(chan Processor),
				Out:             make(chan Processor),
				StorageProvider: &mockStorage{},
			},
			nil,
			false,
			true,
		},
		{
			"No Storage Provider",
			fields{
				In:         make(chan Processor),
				Out:        make(chan Processor),
				TempFolder: "./testdata/tmp",
			},
			nil,
			false,
			true,
		},
		{
			"Valid Item",
			fields{
				In:              make(<-chan Processor),
				Out:             make(chan Processor),
				StorageProvider: &mockStorage{},
				TempFolder:      "./testdata/tmp",
			},
			[]Processor{
				&Info{
					Process: Process{
						Message: message.Message{
							Title: "Test",
							Slug:  "test",
						},
						Result: audit.Result{
							"checksum": "39c7d71a68565ddd7b6a0fd68d94924d0db449a99541439b3ab8a477c5f1fc4e",
						},
					},
				},
			},
			false,
			false,
		},
		{
			"Invalid Item - Checksum",
			fields{
				In:              make(<-chan Processor),
				Out:             make(chan Processor),
				StorageProvider: &mockStorage{},
				TempFolder:      "./testdata/tmp",
			},
			[]Processor{
				&Info{
					Process: Process{
						Message: message.Message{
							Title: "Test",
							Slug:  "test",
						},
						Result: audit.Result{},
					},
				},
			},
			true,
			false,
		},
		{
			"Invalid Item - File Write Error",
			fields{
				In:              make(<-chan Processor),
				Out:             make(chan Processor),
				StorageProvider: &mockStorage{},
				TempFolder:      "./testdata/tmp",
			},
			[]Processor{
				&Info{
					Process: Process{
						Message: message.Message{
							Title: "File Error",
							Slug:  "test",
						},
						Result: audit.Result{
							"checksum": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
						},
					},
				},
			},
			true,
			false,
		},
		{
			"Lighthouse Command - Error",
			fields{
				In:              make(<-chan Processor),
				Out:             make(chan Processor),
				StorageProvider: &mockStorage{},
				TempFolder:      "./testdata/tmp",
			},
			[]Processor{
				&Info{
					Process: Process{
						Message: message.Message{
							Title: "LH Error",
							Slug:  "error",
						},
						Result: audit.Result{
							"checksum": "1234567890",
						},
					},
				},
			},
			true,
			false,
		},
		{
			"Lighthouse Command - JSON Error",
			fields{
				In:              make(<-chan Processor),
				Out:             make(chan Processor),
				StorageProvider: &mockStorage{},
				TempFolder:      "./testdata/tmp",
			},
			[]Processor{
				&Info{
					Process: Process{
						Message: message.Message{
							Title: "LH JSON Error",
							Slug:  "jsonError",
						},
						Result: audit.Result{
							"checksum": "1234567890",
						},
					},
				},
			},
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			lh := &Lighthouse{
				Process:         tt.fields.Process,
				In:              tt.fields.In,
				Out:             tt.fields.Out,
				TempFolder:      tt.fields.TempFolder,
				StorageProvider: tt.fields.StorageProvider,
			}

			lh.SetContext(ctx)
			if tt.procs != nil {
				lh.In = generateProcs(ctx, tt.procs)
			}

			var errc <-chan error
			var err error

			go func() {
				errc, err = lh.Run()
			}()

			// Sleep a short time delay to give process time to start.
			time.Sleep(time.Millisecond * 100)

			if (err != nil) != tt.wantErr {
				t.Errorf("Lighthouse.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (len(errc) != 0) != tt.wantErrc {
				e := <-errc
				t.Errorf("Lighthouse.Run() error = %v, wantErrc %v", e, tt.wantErrc)
				return
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

	case "lh":
		switch args[0] {
		case "https://wp-themes.com/error":
			fmt.Fprintf(os.Stderr, "Error occurred.")
			os.Exit(1)
		case "https://wp-themes.com/jsonError":
			fmt.Fprintf(os.Stdout, "Invalid json.")
			os.Exit(0)
		case "https://wp-themes.com/StdoutPipeError":
			fmt.Fprintf(os.Stderr, "Error occurred.")
			os.Exit(1)
		case "https://wp-themes.com/StderrPipeError":
			fmt.Fprintf(os.Stderr, "Error occurred.")
			os.Exit(1)
		case "https://wp-themes.com/StartError":
			fmt.Fprintf(os.Stderr, "Error occurred.")
			os.Exit(1)
		case "https://wp-themes.com/WaitError":
			fmt.Fprintf(os.Stderr, "Error occurred.")
			os.Exit(1)
		default:
			fmt.Fprintf(os.Stdout, exampleReport())
			os.Exit(0)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %q\n", cmd)
		os.Exit(2)
	}
}

func exampleReport() string {
	return `{
  "reportCategories": [
    {
      "name": "Performance",
      "description": "These encapsulate your web app's current performance and opportunities to improve it.",
      "id": "performance",
      "score": 72.17647058823529
    },
    {
      "name": "Progressive Web App",
      "weight": 1,
      "description": "These checks validate the aspects of a Progressive Web App, as specified by the baseline [PWA Checklist](https://developers.google.com/web/progressive-web-apps/checklist).",
      "id": "pwa",
      "score": 54.54545454545455
    },
    {
      "name": "Accessibility",
      "description": "These checks highlight opportunities to [improve the accessibility of your web app](https://developers.google.com/web/fundamentals/accessibility). Only a subset of accessibility issues can be automatically detected so manual testing is also encouraged.",
      "id": "accessibility",
      "score": 100
    },
    {
      "name": "Best Practices",
      "description": "We've compiled some recommendations for modernizing your web app and avoiding performance pitfalls.",
      "id": "best-practices",
      "score": 81.25
    },
    {
      "name": "SEO",
      "description": "These checks ensure that your page is optimized for search engine results ranking. There are additional factors Lighthouse does not check that may affect your search ranking. [Learn more](https://support.google.com/webmasters/answer/35769).",
      "id": "seo",
      "score": 90
    }
  ]
}`
}
