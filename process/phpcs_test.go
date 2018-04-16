package process

import (
	"testing"
	"context"
	"time"
	"github.com/wptide/pkg/storage"
	"github.com/wptide/pkg/message"
	"errors"
	"github.com/wptide/pkg/shell"
	"os"
	"strings"
	"io/ioutil"
	"bytes"
	"github.com/wptide/pkg/log"
)

type mockPhpcsRunner struct{}

func (m mockPhpcsRunner) Run(name string, arg ...string) ([]byte, []byte, error, int) {

	// "--basepath="
	basepath := strings.Split(arg[4], "=")[1]
	standard := strings.Split(arg[2], "=")[1]

	if basepath == "./testdata/info/plugin/unzipped" && standard == "wordpress" {
		// Simulate phpcs report written to tmp file.
		data := examplePhpcsWordPressReport()
		ioutil.WriteFile(
			"./testdata/tmp/39c7d71a68565ddd7b6a0fd68d94924d0db449a99541439b3ab8a477c5f1fc4e-phpcs_wordpress-full.json",
			[]byte(data),
			0644,
		)

		return []byte("[TEST] Time: 100ms; Memory: 4Mb"), nil, nil, 0
	}

	if basepath == "./testdata/info/plugin/unzipped" && standard == "phpcompatibility" {
		// Simulate phpcs report written to tmp file.
		data := examplePhpcsPhpCompatibilityReport()
		ioutil.WriteFile(
			"./testdata/tmp/39c7d71a68565ddd7b6a0fd68d94924d0db449a99541439b3ab8a477c5f1fc4e-phpcs_phpcompatibility-full.json",
			[]byte(data),
			0644,
		)

		return []byte("[TEST] Time: 50ms; Memory: 4Mb"), nil, nil, 0
	}

	if basepath == "./testdata/info/filereadererror/unzipped" {
		msg := "this is not json!"
		ioutil.WriteFile(
			"./testdata/tmp/filereadererror-phpcs_wordpress-full.json",
			[]byte(msg),
			os.ModePerm,
		)
		return []byte(msg), nil, nil, 0
	}

	if basepath == "./testdata/info/phpcompatwriteerror/unzipped" {
		msg := examplePhpcsPhpCompatibilityReport()
		ioutil.WriteFile(
			"./testdata/tmp/phpcompatwriteerror-phpcs_phpcompatibility-full.json",
			[]byte(msg),
			os.ModePerm,
		)
		return []byte(msg), nil, nil, 0
	}

	if basepath == "./testdata/info/phpcompatuploaderror/unzipped" {
		msg := examplePhpcsPhpCompatibilityReport()
		ioutil.WriteFile(
			"./testdata/tmp/phpcompatuploaderror-phpcs_phpcompatibility-full.json",
			[]byte(msg),
			os.ModePerm,
		)
		return []byte(msg), nil, nil, 0
	}

	return nil, nil, errors.New("Something went wrong."), 1
}

func mockOpen(name string) (*os.File, error) {
	if strings.Contains(name, "39c7d71a68565ddd7b6a0fd68d94924d0db449a99541439b3ab8a477c5f1fc4e") {
		name = strings.Replace(name, "/tmp/", "/results/", -1)
	}

	return os.Open(name)
}

func TestPhpcs_Run(t *testing.T) {

	b := bytes.Buffer{}
	log.SetOutput(&b)
	defer log.SetOutput(os.Stdout)

	// Set out execCommand variable to the mock function.
	runner = &mockPhpcsRunner{}
	// Remember to set it back after the test.
	defer func() { runner = &shell.Command{} }()

	// Set out execCommand variable to the mock function.
	writeFile = mockWriteFile
	// Remember to set it back after the test.
	defer func() { writeFile = ioutil.WriteFile }()

	// Set open file to mock function.
	fileOpen = mockOpen
	// Set it back to os.Open
	defer func() { fileOpen = os.Open }()

	// Need to test with a context.
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	// Make temp folder and clean.
	os.MkdirAll("./testdata/tmp", os.ModePerm)
	defer os.RemoveAll("./testdata/tmp")

	// Make upload folder and clean.
	os.MkdirAll("./testdata/upload", os.ModePerm)
	defer os.RemoveAll("./testdata/upload")

	auditsWordPress := &[]message.Audit{
		{
			Type: "phpcs",
			Options: &message.AuditOption{
				Standard: "wordpress",
			},
		},
	}

	auditsPhpCompatibility := &[]message.Audit{
		{
			Type: "phpcs",
			Options: &message.AuditOption{
				Standard:   "phpcompatibility",
				RuntimeSet: "testVersion 5.2-",
			},
		},
	}

	auditsBoth := &[]message.Audit{
		{
			Type: "phpcs",
			Options: &message.AuditOption{
				Standard: "wordpress",
			},
		},
		{
			Type: "phpcs",
			Options: &message.AuditOption{
				Standard: "phpcompatibility",
			},
		},
	}

	auditsInvalidStandard := &[]message.Audit{
		{
			Type:    "phpcs",
			Options: &message.AuditOption{},
		},
	}

	type fields struct {
		Process         Process
		In              <-chan Processor
		Out             chan Processor
		TempFolder      string
		StorageProvider storage.StorageProvider
	}

	validFields := fields{
		In:              make(<-chan Processor),
		Out:             make(chan Processor),
		StorageProvider: &mockStorage{},
		TempFolder:      "./testdata/tmp",
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
			"Valid Item - WordPress",
			validFields,
			[]Processor{
				&Info{
					Process: Process{
						Message: message.Message{
							Title:  "Valid Test",
							Slug:   "test",
							Audits: auditsWordPress,
						},
						Result: map[string]interface{}{
							"checksum": "39c7d71a68565ddd7b6a0fd68d94924d0db449a99541439b3ab8a477c5f1fc4e",
						},
						FilesPath: "./testdata/info/plugin",
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
							Title:  "Checksum Test",
							Slug:   "test",
							Audits: auditsWordPress,
						},
						Result: make(map[string]interface{}),
					},
				},
			},
			true,
			false,
		},
		{
			"Invalid Item - Standard",
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
							Title:  "Standards Test",
							Slug:   "test",
							Audits: auditsInvalidStandard,
						},
						Result: make(map[string]interface{}),
					},
				},
			},
			true,
			false,
		},
		{
			"Invalid Item - Standard 2",
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
							Title:  "Standards Test",
							Slug:   "test",
							Audits: auditsWordPress,
						},
						Result: map[string]interface{}{
							"checksum": "39c7d71a68565ddd7b6a0fd68d94924d0db449a99541439b3ab8a477c5f1fc4e",
						},
						FilesPath: "",
					},
				},
			},
			true,
			false,
		},
		{
			"Valid Item - Phpcompatibility",
			validFields,
			[]Processor{
				&Info{
					Process: Process{
						Message: message.Message{
							Title:  "Valid Phpcompat",
							Slug:   "test",
							Audits: auditsPhpCompatibility,
						},
						Result: map[string]interface{}{
							"checksum": "39c7d71a68565ddd7b6a0fd68d94924d0db449a99541439b3ab8a477c5f1fc4e",
						},
						FilesPath: "./testdata/info/plugin",
					},
				},
			},
			false,
			false,
		},
		{
			"Valid Item - Multiple",
			validFields,
			[]Processor{
				&Info{
					Process: Process{
						Message: message.Message{
							Title:  "Multiple Standards",
							Slug:   "test",
							Audits: auditsBoth,
						},
						Result: map[string]interface{}{
							"checksum": "39c7d71a68565ddd7b6a0fd68d94924d0db449a99541439b3ab8a477c5f1fc4e",
						},
						FilesPath: "./testdata/info/plugin",
					},
				},
			},
			false,
			false,
		},
		{
			"Not PHPCS",
			validFields,
			[]Processor{
				&Info{
					Process: Process{
						Message: message.Message{
							Title: "Not PHPCS",
							Slug:  "Not PHPCS",
							Audits: &[]message.Audit{
								{
									Type: "lighthouse",
								},
							},
						},
						Result: map[string]interface{}{
							"checksum": "1234567890",
						},
					},
				},
			},
			false,
			false,
		},
		{
			"Upload Error",
			validFields,
			[]Processor{
				&Info{
					Process: Process{
						Message: message.Message{
							Title:  "Upload Error",
							Slug:   "test",
							Audits: auditsWordPress,
						},
						Result: map[string]interface{}{
							"checksum": "uploaderrorchecksum",
						},
						FilesPath: "./testdata/info/plugin",
					},
				},
			},
			true,
			false,
		},
		{
			"Close Context",
			fields{
				In:              make(<-chan Processor),
				Out:             make(chan Processor),
				StorageProvider: &mockStorage{},
				TempFolder:      "closeContext",
			},
			[]Processor{},
			false,
			false,
		},
		{
			"Invalid - JSON Error",
			validFields,
			[]Processor{
				&Info{
					Process: Process{
						Message: message.Message{
							Title:  "JSON Error Test",
							Slug:   "test",
							Audits: auditsWordPress,
						},
						Result: map[string]interface{}{
							"checksum": "filereadererror",
						},
						FilesPath: "./testdata/info/filereadererror",
					},
				},
			},
			true,
			false,
		},
		{
			"phpcompatibility - Report write error",
			validFields,
			[]Processor{
				&Info{
					Process: Process{
						Message: message.Message{
							Title:  "phpcompat report error",
							Slug:   "test",
							Audits: auditsPhpCompatibility,
						},
						Result: map[string]interface{}{
							"checksum": "phpcompatwriteerror",
						},
						FilesPath: "./testdata/info/phpcompatwriteerror",
					},
				},
			},
			true,
			false,
		},
		{
			"phpcompatibility - Report upload error",
			validFields,
			[]Processor{
				&Info{
					Process: Process{
						Message: message.Message{
							Title:  "phpcompat upload error",
							Slug:   "test",
							Audits: auditsPhpCompatibility,
						},
						Result: map[string]interface{}{
							"checksum": "phpcompatuploaderror",
						},
						FilesPath: "./testdata/info/phpcompatuploaderror",
					},
				},
			},
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &Phpcs{
				Process:         tt.fields.Process,
				In:              tt.fields.In,
				Out:             tt.fields.Out,
				TempFolder:      tt.fields.TempFolder,
				StorageProvider: tt.fields.StorageProvider,
			}

			cs.SetContext(ctx)
			if tt.procs != nil {
				cs.In = generateProcs(ctx, tt.procs)
			}

			var errc <-chan error
			var err error

			go func() {
				errc, err = cs.Run()
			}()

			// Sleep a short time delay to give process time to start.
			time.Sleep(time.Millisecond * 100)

			if tt.wantErrc {
				time.Sleep(time.Millisecond * 500)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Phpcs.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (len(errc) != 0) && ! tt.wantErrc {
				e := <-errc
				t.Errorf("Phpcs.Run() error = %v, wantErrc %v", e, tt.wantErrc)
				return
			}
		})
	}
}

func examplePhpcsWordPressReport() string {
	return `{"totals":{"errors":19,"warnings":0,"fixable":12},"files":{"dummy-plugin.php":{"errors":19,"warnings":0,"messages":[{"message":"Class file names should be based on the class name with \"class-\" prepended. Expected class-hello.php, but found dummy-plugin.php.","source":"WordPress.Files.FileName.InvalidClassFileName","severity":5,"type":"ERROR","line":1,"column":1,"fixable":false},{"message":"You must use \"\/**\" style comments for a class comment","source":"Squiz.Commenting.ClassComment.WrongStyle","severity":5,"type":"ERROR","line":35,"column":1,"fixable":false},{"message":"You must use \"\/**\" style comments for a member variable comment","source":"Squiz.Commenting.VariableComment.WrongStyle","severity":5,"type":"ERROR","line":38,"column":13,"fixable":false},{"message":"Tabs must be used to indent lines; spaces are not allowed","source":"Generic.WhiteSpace.DisallowSpaceIndent.SpacesUsed","severity":5,"type":"ERROR","line":40,"column":1,"fixable":true},{"message":"No space after opening parenthesis is prohibited","source":"WordPress.WhiteSpace.ControlStructureSpacing.NoSpaceAfterOpenParenthesis","severity":5,"type":"ERROR","line":41,"column":12,"fixable":true},{"message":"You must use \"\/**\" style comments for a function comment","source":"Squiz.Commenting.FunctionComment.WrongStyle","severity":5,"type":"ERROR","line":41,"column":12,"fixable":false},{"message":"Expected 1 spaces between opening bracket and argument \"$addressee\"; 0 found","source":"Squiz.Functions.FunctionDeclarationArgumentSpacing.SpacingAfterOpen","severity":5,"type":"ERROR","line":41,"column":33,"fixable":true},{"message":"String \"World\" does not require double quotes; use single quotes instead","source":"Squiz.Strings.DoubleQuoteUsage.NotRequired","severity":5,"type":"ERROR","line":41,"column":46,"fixable":true},{"message":"No space before closing parenthesis is prohibited","source":"WordPress.WhiteSpace.ControlStructureSpacing.NoSpaceBeforeCloseParenthesis","severity":5,"type":"ERROR","line":41,"column":53,"fixable":true},{"message":"PHP syntax error: syntax error, unexpected '='","source":"Generic.PHP.Syntax.PHPSyntax","severity":5,"type":"ERROR","line":42,"column":1,"fixable":false},{"message":"Expected 1 space before \"-\"; 0 found","source":"WordPress.WhiteSpace.OperatorSpacing.NoSpaceBefore","severity":5,"type":"ERROR","line":42,"column":14,"fixable":true},{"message":"Expected 1 space after \"-\"; 0 found","source":"WordPress.WhiteSpace.OperatorSpacing.NoSpaceAfter","severity":5,"type":"ERROR","line":42,"column":14,"fixable":true},{"message":"You must use \"\/**\" style comments for a function comment","source":"Squiz.Commenting.FunctionComment.WrongStyle","severity":5,"type":"ERROR","line":46,"column":12,"fixable":false},{"message":"String \"Hello \" does not require double quotes; use single quotes instead","source":"Squiz.Strings.DoubleQuoteUsage.NotRequired","severity":5,"type":"ERROR","line":47,"column":14,"fixable":true},{"message":"Expected next thing to be an escaping function (see Codex for 'Data Validation'), not '$this'","source":"WordPress.XSS.EscapeOutput.OutputNotEscaped","severity":5,"type":"ERROR","line":47,"column":25,"fixable":false},{"message":"Expected 1 spaces after opening bracket; 0 found","source":"PEAR.Functions.FunctionCallSignature.SpaceAfterOpenBracket","severity":5,"type":"ERROR","line":53,"column":16,"fixable":true},{"message":"Expected 1 spaces before closing bracket; 0 found","source":"PEAR.Functions.FunctionCallSignature.SpaceBeforeCloseBracket","severity":5,"type":"ERROR","line":53,"column":16,"fixable":true},{"message":"String \"Mundo\" does not require double quotes; use single quotes instead","source":"Squiz.Strings.DoubleQuoteUsage.NotRequired","severity":5,"type":"ERROR","line":53,"column":22,"fixable":true},{"message":"File must end with a newline character","source":"Generic.Files.EndFileNewline.NotFound","severity":5,"type":"ERROR","line":55,"column":18,"fixable":true}]}}}`
}

func examplePhpcsPhpCompatibilityReport() string {
	return `{"totals":{"errors":4,"warnings":0,"fixable":0},"files":{"phpcompat\/compatissues.php":{"errors":4,"warnings":0,"messages":[{"message":"\"namespace\" keyword is not present in PHP version 5.2 or earlier","source":"PHPCompatibility.PHP.NewKeywords.t_namespaceFound","severity":5,"type":"ERROR","line":3,"column":1,"fixable":false},{"message":"\"trait\" keyword is not present in PHP version 5.3 or earlier","source":"PHPCompatibility.PHP.NewKeywords.t_traitFound","severity":5,"type":"ERROR","line":8,"column":1,"fixable":false},{"message":"Short array syntax (open) is available since 5.4","source":"PHPCompatibility.PHP.ShortArray.Found","severity":5,"type":"ERROR","line":9,"column":9,"fixable":false},{"message":"Short array syntax (close) is available since 5.4","source":"PHPCompatibility.PHP.ShortArray.Found","severity":5,"type":"ERROR","line":9,"column":10,"fixable":false}]},"dummy-plugin.php":{"errors":0,"warnings":0,"messages":[]}}}
`
}
