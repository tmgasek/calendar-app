package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/tmgasek/calendar-app/internal/data"
	"github.com/tmgasek/calendar-app/ui"
)

type templateData struct {
	Form                any
	Flash               string
	IsAuthenticated     bool
	CSRFToken           string
	UserId              int
	Events              []*data.Event
	HourlyAvailability  []HourlyAvailability
	Hours               [16]int
	AppointmentRequests []*data.AppointmentRequest
	Appointments        []*data.Appointment
	User                *data.User
	Users               []*data.User
	Settings            *data.Settings
	TargetUserID        int
	Groups              []*data.Group
	Group               *data.Group
	ErrorData           *ErrorData
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2006 at 15:04")
}

func formatEventTimes(start, end time.Time) string {
	startTime := start.UTC().Format("02 Jan 2006 at 15:04")
	if start.Format("02 Jan 2006") == end.Format("02 Jan 2006") {
		// Same day, only show the end time hour
		return startTime + " - " + end.UTC().Format("15:04")
	} else {
		// Different days, show full end time
		return startTime + " - " + end.UTC().Format("02 Jan 2006 at 15:04")
	}
}

// Init empty funcMap obj and store it in a global var. String keyed map acting
// as a lookup between the names of custom template funcs and actual funcs.
var functions = template.FuncMap{
	"humanDate":        humanDate,
	"formatEventTimes": formatEventTimes,
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
