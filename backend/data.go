package backend

import (
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"
)

var (
	regexIsHttp = regexp.MustCompile(`^https?://`)
)

type dataUsable interface {
	SetData(*data)
}

type data struct {
	title       string
	Article     bool
	Domain      string
	URL         string
	Image       string
	Description string
	Name        string
	Links       []Link
	Logo        *Logo
}

func (d *data) handleGeneric(w http.ResponseWriter, r *http.Request, name string, custom dataUsable) {
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
	if d.Links == nil {
		d.Links = cfg.Links
	}
	if d.Logo == nil {
		d.Logo = &cfg.Logo
	}
	if d.URL == "" {
		if !strings.HasPrefix(r.URL.Path, "/") {
			r.URL.Path = "/" + r.URL.Path
		}
		d.URL = r.URL.Path
	}
	t, err := template.New("").Funcs(template.FuncMap{
		"static": func(path string) string {
			if regexIsHttp.MatchString(path) {
				return path
			}
			return fmt.Sprintf("/static/%s", path)
		},
		"assets": func(path string) string {
			if regexIsHttp.MatchString(path) {
				return path
			}
			return fmt.Sprintf("/assets/%s", path)
		},
		"next":   func(i int) int { return i + 1 },
		"before": func(i int) int { return i - 1 },
	}).ParseFS(templates, fmt.Sprintf("templates/%s.html", name), "templates/base.html")
	if err != nil {
		panic(err)
	}
	if custom == nil {
		err = t.ExecuteTemplate(w, "base.html", d)
	} else {
		custom.SetData(d)
		err = t.ExecuteTemplate(w, "base.html", custom)
	}
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

func (d *data) PubDate() string {
	return ""
}
