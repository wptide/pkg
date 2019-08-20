package tide

import "reflect"

// ResultSet contains results as a slice of Items.
type ResultSet struct {
	Results []Item
}

// Item describes an item in a result.
type Item struct {
	Title         string                 `json:"title"`
	Description   string                 `json:"content"`
	Version       string                 `json:"version"`
	Checksum      string                 `json:"checksum"`
	Visibility    string                 `json:"visibility"`
	ProjectType   string                 `json:"project_type"`
	SourceURL     string                 `json:"source_url"`
	SourceType    string                 `json:"source_type"`
	CodeInfo      CodeInfo               `json:"code_info,omitempty"`
	Reports       map[string]AuditResult `json:"reports,omitempty"`
	Standards     []string               `json:"standards,omitempty"`      // Will potentially be overriden in API and should not be relied upon.
	RequestClient string                 `json:"request_client,omitempty"` // Will be converted to a user.
	Project       []string               `json:"project,omitempty"`        // Has to be an array of string because of how taxonomies work in WordPress.
}

// CodeInfo contains the details about the files being processed.
type CodeInfo struct {
	Type    string                `json:"type"`
	Details []InfoDetails         `json:"details"`
	Cloc    map[string]ClocResult `json:"cloc"`
}

// InfoDetails is a KV pair describing entries in CodeInfo.
type InfoDetails struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// InfoDetailsSimple is the CodeInfo details converted into a simpler struct.
type InfoDetailsSimple struct {
	Name        string
	PluginURI   string
	ThemeURI    string
	Version     string
	Description string
	Author      string
	AuthorURI   string
	TextDomain  string
}

// ClocResult runs the code through the `clock` package to get information about the source.
type ClocResult struct {
	Blank   int `json:"blank"`
	Comment int `json:"comment"`
	Code    int `json:"code"`
	NFiles  int `json:"n_files"`
}

// AuditResult contain results about an audit.
type AuditResult struct {
	Raw                  AuditDetails           `json:"raw,omitempty"`
	Parsed               AuditDetails           `json:"parsed,omitempty"`
	Summary              AuditSummary           `json:"summary,omitempty"`
	CompatibleVersions   []string               `json:"compatible_versions,omitempty"`
	IncompatibleVersions []string               `json:"incompatible_versions,omitempty"`
	Error                string                 `json:"error,omitempty"`
	Extra                map[string]interface{} `json:"extra,omitempty"`
}

// PhpcsResults contains the results from a phpcs audit.
type PhpcsResults struct {
	Totals struct {
		Errors   int `json:"errors,omitempty"`
		Warnings int `json:"warnings,omitempty"`
	} `json:"totals,omitempty"`
	Files map[string]struct {
		Errors   int                 `json:"errors, omitempty"`
		Warnings int                 `json:"warnings,omitempty"`
		Messages []PhpcsFilesMessage `json:"messages,omitempty"`
	} `json:"files,omitempty"`
}

// PhpcsFilesMessage contains individual violation information about a file.
type PhpcsFilesMessage struct {
	Message  string `json:"message"`
	Source   string `json:"source"`
	Severity int    `json:"severity,omitempty"`
	Type     string `json:"type"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Fixable  bool   `json:"fixable"`
}

// AuditDetails contains report information about performed audits.
type AuditDetails struct {
	Type     string `json:"type,omitempty"`
	FileName string `json:"filename,omitempty"`
	Path     string `json:"path,omitempty"`
	*PhpcsResults
	*LighthouseResults
}

// AuditSummary is a proxy struct for `phpcs` and `lighthouse`.
type AuditSummary struct {
	*PhpcsSummary
	*LighthouseSummary
}

// PhpcsSummary is a simplified version of `phpcs` results.
type PhpcsSummary struct {
	Files map[string]struct {
		Errors   int `json:"errors"`
		Warnings int `json:"warnings"`
	} `json:"files,omitempty"`
	FilesCount    int `json:"files_count"`
	ErrorsCount   int `json:"errors_count"`
	WarningsCount int `json:"warnings_count"`
}

// LighthouseResults is a simplified version of `lighthouse` results.
// TODO: Define this later.
type LighthouseResults struct{}

// LighthouseSummary uses only the catagories information from an extensice Lighthouse report.
type LighthouseSummary struct {
	Categories map[string]LighthouseCategory `json:"categories,omitempty"`
}

// LighthouseCategory contains the results for a given category.
type LighthouseCategory struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	ID          string  `json:"id"`
	Score       float32 `json:"score"`
}

// SimplifyCodeDetails converts []InfoDetails into InfoDetailsSimple.
func SimplifyCodeDetails(details []InfoDetails) *InfoDetailsSimple {
	simple := &InfoDetailsSimple{}

	sV := reflect.ValueOf(simple).Elem()

	for _, item := range details {
		sF := sV.FieldByName(item.Key)
		if sF.IsValid() {
			sF.Set(reflect.ValueOf(item.Value))
		}
	}
	return simple
}

// ComplexifyCodeDetails converts InfoDetailsSimple into []InfoDetails.
func ComplexifyCodeDetails(simple *InfoDetailsSimple) []InfoDetails {

	sV := reflect.ValueOf(*simple)

	details := []InfoDetails{}

	for i := 0; i < sV.NumField(); i++ {
		if sV.Field(i).String() != "" {
			details = append(details, InfoDetails{
				sV.Type().Field(i).Name,
				sV.Field(i).String(),
			})
		}
	}

	return details
}
