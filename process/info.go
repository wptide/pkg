package process

import (
	"errors"
	"io/ioutil"
	"github.com/hhatto/gocloc"
	"strings"
	"regexp"
	"fmt"
	"github.com/wptide/pkg/tide"
	"github.com/wptide/pkg/log"
	"github.com/wptide/pkg/message"
	"golang.org/x/text/transform"
)

// Ingest defines the structure for our Ingest process.
type Info struct {
	Process              // Inherits methods from Process.
	In  <-chan Processor // Expects a processor channel as input.
	Out chan Processor   // Send results to an output channel.
}

// Run executes the process in the pipeline.
func (info *Info) Run(errc *chan error) error {

	if info.In == nil {
		return errors.New("requires a previous process")
	}
	if info.Out == nil {
		return errors.New("requires a next process")
	}

	go func() {

		for {
			select {
			case in := <-info.In:

				// Copy Process fields from `in` process.
				info.CopyFields(in)

				// Run the process.
				// If processing produces an error send it up the error channel.
				if err := info.Do(); err != nil {
					// Pass the error up the error channel.
					*errc <- errors.New("Info Error: " + err.Error())
					// continue so that the message doesn't get passed along.
					continue
				}

				// Send process to the out channel.
				info.Out <- info
			}
		}

	}()

	return nil
}

// process runs the actual code for this process.
func (info *Info) Do() error {

	result := *info.Result

	log.Log(info.Message.Title, "Processing CodeInfo")

	// Try to get filesPath from results first.
	if path, ok := result["filesPath"].(string); ok {
		info.SetFilesPath(path)
	}

	if info.GetFilesPath() == "" {
		return errors.New("could not determine files path")
	}

	path := info.GetFilesPath() + "/unzipped"

	cloc, err := getCloc(path)
	if err != nil {
		return err
	}

	projectType, details, _ := getProjectDetails(info.Message, path)

	result["info"] = tide.CodeInfo{
		Type:    projectType,
		Details: details,
		Cloc:    cloc,
	}
	info.Result = &result

	log.Log(info.Message.Title, "Project is `"+projectType+"`")

	return nil
}

// getProjectDetails attempts to get project details from code base.
func getProjectDetails(msg message.Message, path string) (string, []tide.InfoDetails, error) {

	projectType := "other"
	details := []tide.InfoDetails{}

	type header struct {
		projectType string
		details     []tide.InfoDetails
	}

	var extracted []header

	// Get files in root path.
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return "", nil, err
	}

	for _, f := range files {
		projectType, details, err = extractHeader(path + "/" + f.Name())
		if err == nil {
			extracted = append(extracted, header{
				projectType,
				details,
			})
		}
	}

	// No headers found.
	if len(extracted) == 0 {
		return projectType, details, errors.New("not a theme or plugin")
	}

	// A single header found.
	if len(extracted) == 1 {
		return extracted[0].projectType, extracted[0].details, err
	}

	// Multiple headers found, attempt to match the correct one.

	// Attempt to get slug from filename (purely fallback scenario if text domain does not match).
	filenameMatch := ""
	re := regexp.MustCompile(`(?mU)(?P<slug>[^\/\\]+)(\.\d).+$`)
	matches := re.FindStringSubmatch(msg.SourceURL)
	if len(matches) > 2 {
		filenameMatch = matches[1]
	}

	// Range through each header...
	for _, h := range extracted {
		simplified := tide.SimplifyCodeDetails(h.details)

		// ... match message slug to text domain or filename part and match project types.
		if (simplified.TextDomain == msg.Slug || simplified.TextDomain == filenameMatch) &&
			msg.ProjectType == h.projectType {
			return h.projectType, h.details, nil
		}
	}

	// Multiple headers found but could not assert appropriate header.
	return "", nil, errors.New("Multiple headers: Could not assert appropriate header for project.")
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

// cr2nl is a transformer to conver \r endings to \n.
type cr2nl struct{ transform.NopResetter }

// Transform implements transform.Transformer.
func (cr2nl) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	nSrc = copy(dst, src)
	nDst = nSrc
	dst = dst[:nDst]
	for i := range dst {
		if dst[i] == '\r' {
			dst[i] = '\n'
		}
	}
	return nSrc, nDst, nil
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

	f, _ := fileOpen(filename)
	defer f.Close()

	// Pass the `f` reader to a transformed reader.
	tr := transform.NewReader(f, cr2nl{})
	b1 := make([]byte, 8192)
	// Use the transform reader instead of file.
	n1, _ := tr.Read(b1)

	isStyleCSS, _ := regexp.Match(`(\/style.css)$`, []byte(filename))

	if n1 > 0 {

		fileHeader := strings.Replace(string(b1), "\r\n", "\n", -1)

		validHeader := false
		for _, field := range headerFields {
			pattern := fmt.Sprintf("%s:.*", field)
			re := regexp.MustCompile(pattern)
			value := strings.Replace(re.FindString(fileHeader), field+":", "", -1)
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
