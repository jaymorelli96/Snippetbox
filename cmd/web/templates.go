package main

import (
	"io/fs"
	"net/http"
	"path/filepath"
	"text/template"
	"time"

	"github.com/justinas/nosurf"
	"snippetbox.jmorelli.dev/internal/models"
	"snippetbox.jmorelli.dev/ui"
)

type templateData struct {
	CurrentYear     int
	Snippet         *models.Snippet
	Snippets        []*models.Snippet
	User            *models.User
	Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
}

func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		IsAuthenticated: app.isAuthenticated(r),
		CSRFToken:       nosurf.Token(r),
	}
}

func humanDate(time time.Time) string {
	if time.IsZero() {
		return ""
	}

	return time.UTC().Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)

	pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string{
			"html/base.tmpl.html",
			"html/partials/*.tmpl.html",
			page,
		}

		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
