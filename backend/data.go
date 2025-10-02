package backend

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type data struct {
	title       string
	Article     bool
	Domain      string
	URL         string
	Image       string
	Description string
	Name        string
}

func (d *data) handleGeneric(w http.ResponseWriter, r *http.Request, name string) {
	cfg := r.Context().Value("config").(*Config)
	if d.Domain == "" {
		d.Domain = cfg.Domain
	}
	if d.Name == "" {
		d.Name = cfg.Name
	}
	if d.Description == "" {
		d.Description = cfg.Description
	}
	if d.URL == "" {
		if !strings.HasPrefix(r.URL.Path, "/") {
			r.URL.Path = "/" + r.URL.Path
		}
		d.URL = r.URL.Path
	}
	t, err := template.New("").Funcs(template.FuncMap{
		"static": func(path string) string {
			return fmt.Sprintf("/static/%s", path)
		},
		"assets": func(path string) string {
			return fmt.Sprintf("/assets/%s", path)
		},
	}).ParseFS(templates, fmt.Sprintf("templates/%s.html", name), "templates/base.html")
	if err != nil {
		panic(err)
	}
	err = t.ExecuteTemplate(w, "base.html", d)
	if err != nil {
		panic(err)
	}
}

func (d *data) Title() string {
	title := d.Name
	if d.Article {
		title += " - log entry"
	}
	if len(d.title) != 0 {
		title += " - " + d.title
	}
	return title
}
