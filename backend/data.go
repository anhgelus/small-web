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
}

func (d *data) handleGeneric(w http.ResponseWriter, r *http.Request, name string) {
	if d.Domain == "" {
		cfg := r.Context().Value("config").(*Config)
		d.Domain = cfg.Domain
	}
	if d.URL == "" {
		d.URL = strings.TrimPrefix(r.URL.Path, "/")
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
	title := "anhgelus"
	if d.Article {
		title += " - log entry"
	}
	if len(d.title) != 0 {
		title += " - " + d.title
	}
	return title
}
