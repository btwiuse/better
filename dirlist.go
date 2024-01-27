package better

import (
	"embed"
	"html/template"
	"net/http"
	"strings"
	"time"
)

// Assets holds the embedded filesystem for the default template
//
//go:embed templates
var Assets embed.FS

func dirList(w http.ResponseWriter, r *http.Request, dir http.File) {
	remote := strings.TrimPrefix(r.URL.Path, "/")

	stat, err := dir.Stat()
	if err != nil {
		return
	}

	dirEntries, err := dir.Readdir(0)
	if err != nil {
		// serve.Error(s, w, "Failed to list directory", err)
		return
	}

	tmpl, err := GetTemplate()
	if err != nil {
		return
	}

	// Make the entries for display
	directory := NewDirectory(remote, tmpl)
	for _, node := range dirEntries {
		directory.AddHTMLEntry(node.Name(), node.IsDir(), node.Size(), node.ModTime().UTC())
	}

	sortParm := r.URL.Query().Get("sort")
	orderParm := r.URL.Query().Get("order")
	directory.ProcessQueryParams(sortParm, orderParm)

	// Set the Last-Modified header to the timestamp
	w.Header().Set("Last-Modified", stat.ModTime().UTC().Format(http.TimeFormat))

	directory.Serve(w, r)
}

// GetTemplate returns the HTML template for serving directories via HTTP/WebDAV
func GetTemplate() (*template.Template, error) {
	tmpl := "templates/index.html"

	data, err := Assets.ReadFile(tmpl)
	if err != nil {
		return nil, err
	}

	funcMap := template.FuncMap{
		"afterEpoch": AfterEpoch,
		"contains":   strings.Contains,
		"hasPrefix":  strings.HasPrefix,
		"hasSuffix":  strings.HasSuffix,
	}

	tpl, err := template.New("index").Funcs(funcMap).Parse(string(data))
	if err != nil {
		return nil, err
	}

	return tpl, nil
}

// AfterEpoch returns the time since the epoch for the given time
func AfterEpoch(t time.Time) bool {
	return t.After(time.Time{})
}
