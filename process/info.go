package process

import (
	"errors"
	"io/ioutil"
	"github.com/hhatto/gocloc"
	"strings"
	"os"
	"regexp"
	"fmt"
	"github.com/wptide/pkg/tide"
	"github.com/wptide/pkg/log"
)

// Ingest defines the structure for our Ingest process.
type Info struct {
	Process              // Inherits methods from Process.
	In  <-chan Processor // Expects a processor channel as input.
	Out chan Processor   // Send results to an output channel.
}

// Run executes the process in the pipeline.
func (info *Info) Run() (<-chan error, error) {

	if info.In == nil {
		return nil, errors.New("requires a previous process")
	}
	if info.Out == nil {
		return nil, errors.New("requires a next process")
	}

	errc := make(chan error, 1)

	go func() {
		defer close(errc)

		for {
			select {
			case in := <-info.In:

				// Copy Process fields from `in` process.
				info.CopyFields(in)

				// Run the process.
				// If processing produces an error send it up the error channel.
				if err := info.process(); err != nil {
					// Pass the error up the error channel.
					errc <- err
					break
				}

				// Send process to the out channel.
				info.Out <- info
			case <-info.context.Done():
				return
			}
		}

	}()

	return errc, nil
}

// process runs the actual code for this process.
func (info *Info) process() error {

	log.Log(info.Message.Title, "Processing CodeInfo")

	if info.GetFilesPath() == "" {
		return errors.New("could not determine files path")
	}

	path := info.GetFilesPath() + "/unzipped"
	//
	cloc, err := getCloc(path)
	if err != nil {
		return err
	}

	projectType, details, _ := getProjectDetails(path)

	info.Result["info"] = &tide.CodeInfo{
		Type:    projectType,
		Details: details,
		Cloc:    cloc,
	}

	log.Log(info.Message.Title, "Project is `"+projectType+"`")

	return nil
}

// getProjectDetails attempts to get project details from code base.
func getProjectDetails(path string) (string, []tide.InfoDetails, error) {

	projectType := "other"
	details := []tide.InfoDetails{}

	var found bool

	// Get files in root path.
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return "", nil, err
	}

	// Traverse files and scan for headers.
	for _, f := range files {
		projectType, details, err = extractHeader(path + "/" + f.Name())
		if err == nil {
			found = true
			break
		}
	}

	if ! found {
		err = errors.New("not a theme or plugin")
	}

	return projectType, details, err
}

// getCloc gets the code info for the current code base.
func getCloc(path string) (map[string]tide.ClocResult, error) {

	clocMap := make(map[string]tide.ClocResult)

	languages := gocloc.NewDefinedLanguages()
	options := gocloc.NewClocOptions()
	paths := []string{path}

	processor := gocloc.NewProcessor(languages, options)
	cloc, err := processor.Analyze(paths)

	if err != nil {
		return nil, err
	}

	clocTotals := tide.ClocResult{0, 0, 0, 0}

	for _, cLang := range cloc.Languages {
		// Add Totals
		clocTotals.Code += int(cLang.Code)
		clocTotals.Blank += int(cLang.Blanks)
		clocTotals.Comment += int(cLang.Comments)
		clocTotals.NFiles += len(cLang.Files)

		clocMap[strings.ToLower(cLang.Name)] = tide.ClocResult{
			int(cLang.Blanks),
			int(cLang.Comments),
			int(cLang.Code),
			len(cLang.Files),
		}
	}

	clocMap["sum"] = clocTotals

	return clocMap, nil
}

// extractHeader scans every .php file in the path to retrieve a possible plugin header, or
// looks for style.css to extract the theme header.
//
// The information is return as an interface map.
func extractHeader(filename string) (projectType string, details []tide.InfoDetails, err error) {

	projectType = "other"

	headerFields := []string{
		"Plugin Name",
		"Plugin URI",
		"Description",
		"Version",
		"Author",
		"Author URI",
		"Text Domain",
		"License",
		"License URI",
		"Theme Name",
		"Theme URI",
		"Tags",
	}

	f, _ := os.Open(filename)
	defer f.Close()
	b1 := make([]byte, 8192)
	n1, _ := f.Read(b1)

	isStyleCSS, _ := regexp.Match(`(style.css)$`, []byte(filename))

	if n1 > 0 {

		validHeader := false
		for _, field := range headerFields {
			pattern := fmt.Sprintf("%s:.*", field)
			re := regexp.MustCompile(pattern)
			value := strings.Replace(re.FindString(string(b1)), field+":", "", -1)
			if len(value) > 0 {

				fieldname := field

				if field == "Plugin Name" {
					projectType = "plugin"
					validHeader = true
					fieldname = "Name"
				}
				if field == "Theme Name" && isStyleCSS {
					projectType = "theme"
					validHeader = true
					fieldname = "Name"
				}

				details = append(details, tide.InfoDetails{
					strings.Replace(fieldname, " ", "", -1),
					strings.TrimSpace(value),
				})
			}
		}

		if validHeader {
			return
		}
	}

	return "other", nil, errors.New("not a valid theme or plugin file")
}
