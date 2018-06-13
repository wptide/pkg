package process

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/wptide/pkg/log"
	"github.com/wptide/pkg/message"
	"github.com/wptide/pkg/shell"
	"github.com/wptide/pkg/storage"
)

type mockRunner struct{}

func (m mockRunner) Run(name string, arg ...string) ([]byte, []byte, int, error) {
	switch arg[0] {
	case "https://wp-themes.com/test":
		return []byte(exampleLighthouseReport()), nil, 0, nil
	case "https://wp-themes.com/jsonError":
		return []byte("this is not json"), nil, 0, nil
	case "https://wp-themes.com/error":
		return nil, []byte("error output"), 0, nil
	default:
		return nil, nil, 1, errors.New("something went wrong")
	}
}

func mockWriteFile(filename string, data []byte, perm os.FileMode) error {

	switch filename {
	case "./testdata/tmp/phpcompatwriteerror-phpcs_phpcompatibility-parsed.json":
		fallthrough
	case "./testdata/tmp/ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff-lighthouse-raw.json":
		return errors.New("something went wrong")
	default:
		return ioutil.WriteFile(filename, data, perm)
	}
}

func TestLighthouse_Run(t *testing.T) {

	b := bytes.Buffer{}
	log.SetOutput(&b)
	defer log.SetOutput(os.Stdout)

	// Set out execCommand variable to the mock function.
	lhRunner = &mockRunner{}
	defaultRunner = &mockRunner{}
	// Remember to set it back after the test.
	defer func() {
		lhRunner = &shell.Command{}
		defaultRunner = &shell.Command{}
	}()

	// Set out execCommand variable to the mock function.
	writeFile = mockWriteFile
	// Remember to set it back after the test.
	defer func() { writeFile = ioutil.WriteFile }()

	// Need to test with a context.
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	// Make temp folder and clean.
	os.MkdirAll("./testdata/tmp", os.ModePerm)
	defer os.RemoveAll("./testdata/tmp")

	// Make upload folder and clean.
	os.MkdirAll("./testdata/upload", os.ModePerm)
	defer os.RemoveAll("./testdata/upload")

	audits := []*message.Audit{
		{
			Type: "lighthouse",
		},
	}

	type fields struct {
		Process         Process
		In              <-chan Processor
		Out             chan Processor
		TempFolder      string
		StorageProvider storage.Provider
	}
	tests := []struct {
		name       string
		fields     fields
		procs      []Processor
		mockRunner bool
		wantErrc   bool
		wantErr    bool
	}{
		{
			"Invalid In channel",
			fields{
				Out:             make(chan Processor),
				StorageProvider: &mockStorage{},
				TempFolder:      "./testdata/tmp",
			},
			nil,
			true,
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
			true,
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
			true,
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
			true,
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
							Title:  "Test",
							Slug:   "test",
							Audits: audits,
						},
						Result: &Result{
							"checksum": "39c7d71a68565ddd7b6a0fd68d94924d0db449a99541439b3ab8a477c5f1fc4e",
						},
					},
				},
			},
			true,
			false,
			false,
		},
		{
			"Invalid Message",
			fields{
				In:              make(<-chan Processor),
				Out:             make(chan Processor),
				StorageProvider: &mockStorage{},
				TempFolder:      "./testdata/tmp",
			},
			[]Processor{
				&Info{
					Process: Process{
						Message: message.Message{},
						Result: &Result{
							"checksum": "39c7d71a68565ddd7b6a0fd68d94924d0db449a99541439b3ab8a477c5f1fc4e",
						},
					},
				},
			},
			true,
			true,
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
							Title:  "Test",
							Slug:   "test",
							Audits: audits,
						},
						Result: &Result{},
					},
				},
			},
			true,
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
							Title:  "File Error",
							Slug:   "test",
							Audits: audits,
						},
						Result: &Result{
							"checksum": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
						},
					},
				},
			},
			true,
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
							Title:  "LH Error",
							Slug:   "error",
							Audits: audits,
						},
						Result: &Result{
							"checksum": "1234567890",
						},
					},
				},
			},
			true,
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
							Title:  "LH JSON Error",
							Slug:   "jsonError",
							Audits: audits,
						},
						Result: &Result{
							"checksum": "1234567890",
						},
					},
				},
			},
			true,
			true,
			false,
		},
		{
			"Not Lighthouse",
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
							Title: "Not Lighthouse",
							Slug:  "Not Lighthouse",
							Audits: []*message.Audit{
								{
									Type: "phpcs",
								},
							},
						},
						Result: &Result{
							"checksum": "1234567890",
						},
					},
				},
			},
			true,
			false,
			false,
		},
		{
			"No Temp Folder - No mock runner",
			fields{
				In:              make(chan Processor),
				Out:             make(chan Processor),
				StorageProvider: &mockStorage{},
			},
			nil,
			false,
			false,
			true,
		},
		{
			"Valid Item - Use default runner",
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
							Title:  "Test",
							Slug:   "test",
							Audits: audits,
						},
						Result: &Result{
							"checksum": "39c7d71a68565ddd7b6a0fd68d94924d0db449a99541439b3ab8a477c5f1fc4e",
						},
					},
				},
			},
			false,
			false,
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

			if !tt.mockRunner {
				oldRunner := lhRunner
				lhRunner = nil
				defer func() {
					lhRunner = oldRunner
				}()
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
				err = lh.Run(&errc)
			}()

			// Sleep a short time delay to give process time to start.
			time.Sleep(time.Millisecond * 100)

			if (err != nil) != tt.wantErr {
				t.Errorf("Lighthouse.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (chanError != nil) != tt.wantErrc {
				t.Errorf("Lighthouse.Run() errorChan = %v, wantErrc %v", chanError, tt.wantErrc)
			}
		})
	}
}

func exampleLighthouseReport() string {
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
