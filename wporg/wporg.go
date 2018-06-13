package wporg

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

/*
 *	TimeFormat is the Go-style time format used for plugins.
 *	TimeFormatThemes is the Go-style time format used for themes.
 */
const (
	TimeFormat       = "2006-01-02 3:04pm MST"
	TimeFormatThemes = "2006-01-02"
)

var (
	themesAPIURL  = "https://api.wordpress.org/themes/info/1.1/"
	pluginsAPIURL = "https://api.wordpress.org/plugins/info/1.1/"
)

// APIInfo contains the results from a call to the WordPress.org theme/plugin API.
type APIInfo struct {
	Page    int `json:"page"`
	Pages   int `json:"pages"`
	Results int `json:"results"`
}

// RepoProject describes a single item in the returned results.
type RepoProject struct {
	Name             string `json:"name"`
	Slug             string `json:"slug"`
	Version          string `json:"version"`
	LastUpdated      string `json:"last_updated"`
	ShortDescription string `json:"short_description"`
	Description      string `json:"description"`
	DownloadLink     string `json:"download_link"`
	// Type is excluded from the json object as this does not come
	// from either the plugin or theme APIs.
	Type string `json:"-"`
}

// UnmarshalJSON is a custom unmarshaller for RepoProject to deal with version passed as numbers.
func (rp *RepoProject) UnmarshalJSON(d []byte) error {
	type altRepoProject RepoProject // Do this to remove methods and avoid UnmarshalJSON loop.
	temp := struct {
		altRepoProject
		// Override the version.
		Version json.Number `json:"version"`
	}{
		// Pass in original project pointer so that other fields are not skipped.
		altRepoProject: altRepoProject(*rp),
	}

	// Attempt to unmarshal into the new struct.
	if err := json.Unmarshal(d, &temp); err != nil {
		return err
	}

	// Pass it back to the original.
	*rp = RepoProject(temp.altRepoProject)
	rp.Version = temp.Version.String() // convert the numeric version into string version/

	return nil
}

// APIResponse describes a reponse from the WordPress.org theme/plugin API.
type APIResponse struct {
	Info APIInfo `json:"info"`
	// Type is excluded from the json object as this does not come
	// from either the plugin or theme APIs.
	Type    string        `json:"-"`
	Plugins []RepoProject `json:"plugins,omitempty"`
	Themes  []RepoProject `json:"themes,omitempty"`
}

// alternateAPIResponse is identical to APIResponse except for the
// Plugins field which expects an alternate json response from the
// plugins API. (some inconsistencies with responses were noticed)
type alternateAPIResponse struct {
	Info    APIInfo                `json:"info"`
	Type    string                 `json:"-"`
	Plugins map[string]RepoProject `json:"plugins,omitempty"`
	Themes  []RepoProject          `json:"themes,omitempty"`
}

// Requester describes an interface for requesting projects from an API.
// Client below implements this interface.
type Requester interface {
	// Request performs the request to the WordPress.org API's.
	// `source` is likely to be:
	// - "https://api.wordpress.org/themes/info/1.1/" for themes.
	// - "https://api.wordpress.org/plugins/info/1.1/" for plugins.
	//  `projectType` should be plural "themes" or "plugins".
	Request(source, projectType, category string, perPage, page int) (*APIResponse, error)
}

// Client is the default wporg client and implements Requester interface.
type Client struct {
	pluginAPI string
	themeAPI  string
}

// Request gets information from the WordPress.org API's.
func (c Client) Request(source, projectType, category string, perPage, page int) (*APIResponse, error) {
	// formValues is an array of query parameters passed to the requestUrl.
	formValues := []string{
		"action=query_" + projectType,
		"request[browse]=" + category, // Can be `popular`, `featured`, `updated`, `new`
		"request[per_page]=" + strconv.Itoa(perPage),
		"request[page]=" + strconv.Itoa(page),
		"request[fields][short_description]=1", // Not available for themes.
		"request[fields][sections]=0",
		"request[fields][tested]=0",
		"request[fields][requires]=0",
		"request[fields][rating]=0",
		"request[fields][ratings]=0",
		"request[fields][downloaded]=0",
		"request[fields][last_updated]=1",
		"request[fields][homepage]=0",
		"request[fields][tags]=0",
		"request[fields][donate_link]=0",
		"request[fields][contributors]=0",
		"request[fields][compatibility]=0",
		"request[fields][versions]=0",
		"request[fields][version]=1",
		"request[fields][screenshots]=0",
		"request[fields][stable_tag]=0",
		"request[fields][download_link]=1", // Not available for themes.
		"request[fields][requires_php]=0",
	}
	if projectType == "plugins" {
		// Descriptions are too long for plugins.
		formValues = append(formValues, "request[fields][description]=0")
	} else {
		formValues = append(formValues, "request[fields][description]=1")
	}

	requestParams := bytes.NewBufferString(strings.Join(formValues, "&"))

	response, err := http.Post(source, "application/x-www-form-urlencoded", requestParams)
	if err != nil {
		return nil, errors.New("could not retrieve projects from " + source)
	}

	defer response.Body.Close()
	bodyByte, _ := ioutil.ReadAll(response.Body)

	results := APIResponse{}
	if err := json.Unmarshal([]byte(bodyByte), &results); err != nil {
		// Try alternate form.
		alternate := alternateAPIResponse{}
		if err := json.Unmarshal([]byte(bodyByte), &alternate); err != nil {
			return nil, err
		}
		results.Info = alternate.Info
		results.Themes = alternate.Themes
		plugins := []RepoProject{}
		for _, plugin := range alternate.Plugins {
			plugins = append(plugins, plugin)
		}
		results.Plugins = plugins
	}

	results.Type = projectType

	return &results, nil
}

// RequestThemes is a convenience method.
func (c *Client) RequestThemes(category string, perPage, page int) (*APIResponse, error) {
	if c.themeAPI == "" {
		c.themeAPI = themesAPIURL
	}
	return c.Request(c.themeAPI, "themes", category, perPage, page)
}

// RequestPlugins is a convenience method.
func (c *Client) RequestPlugins(category string, perPage, page int) (*APIResponse, error) {
	if c.pluginAPI == "" {
		c.pluginAPI = pluginsAPIURL
	}
	return c.Request(c.pluginAPI, "plugins", category, perPage, page)
}

// SetPluginAPISource allows to set an alternate plugins API source.
func (c *Client) SetPluginAPISource(source string) {
	c.pluginAPI = source
}

// SetThemeAPISource allows to set an alternate themes API source.
func (c *Client) SetThemeAPISource(source string) {
	c.themeAPI = source
}
