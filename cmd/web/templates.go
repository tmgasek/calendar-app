package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/tmgasek/calendar-app/ui"
)

type templateData struct {
	// Snippet         *models.Snippet
	// Snippets        []*models.Snippet
	CurrentYear     int
	Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2006 at 15:04")
}

// Init empty funcMap obj and store it in a global var. String keyed map acting
// as a lookup between the names of custom template funcs and actual funcs.
var functions = template.FuncMap{
	"humanDate": humanDate,
}

// Only parse files once when app starts, then store the parsed templates in
// an in memory cache.
func newTemplateCache() (map[string]*template.Template, error) {
	// Init new map to act as the cache.
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		// Filepath patterns for the templates we want to parse.
		patterns := []string{
			"html/base.tmpl",
			"html/partials/*.tmpl",
			page,
		}

		// Parse template files from ui.Files embedded filesystem
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
