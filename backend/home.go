package backend

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"git.anhgelus.world/anhgelus/small-world/markdown"
	"github.com/go-chi/chi/v5"
)

const maxLogsPerPage = 5

var (
	sortedLogs  []*logData
	rootContent = map[string]template.HTML{}
)

type homeData struct {
	*data
	Logs        []*logData
	PagesNumber int
	CurrentPage int
}

func (h *homeData) SetData(d *data) {
	h.data = d
}

func HandleHome(r *chi.Mux) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		d := handleGenericLogsDisplay(w, r)
		if d == nil {
			return
		}
		d.handleGeneric(w, r, "home", d)
	})
}

func Handle404(r *chi.Mux) {
	r.NotFound(notFound)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	d := new(data)
	d.title = "404"
	w.WriteHeader(http.StatusNotFound)
	d.handleGeneric(w, r, "404", d)
}

type rootData struct {
	*data
	Content template.HTML
}

func (l *rootData) SetData(d *data) {
	l.data = d
}

func HandleRoot(r *chi.Mux, cfg *Config) {
	err := os.Mkdir(cfg.RootFolder, 0774)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
	r.Get("/{name:[a-zA-Z-]+}", func(w http.ResponseWriter, r *http.Request) {
		handleGenericRoot(w, r, chi.URLParam(r, "name"))
	})
}

func handleGenericRoot(w http.ResponseWriter, r *http.Request, name string) {
	d := new(rootData)
	d.data = new(data)
	if c, ok := rootContent[name]; ok {
		d.Content = c
	} else {
		cfg := r.Context().Value(configKey).(*Config)
		path := filepath.Join(cfg.RootFolder, name+".md")
		b, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				notFound(w, r)
				return
			}
			panic(err)
		}
		var errMd *markdown.ParseError
		d.Content, errMd = markdown.ParseBytes(b, &markdown.Option{ImageSource: getStatic})
		if errMd != nil {
			slog.Error("parsing markdown", "path", path)
			fmt.Println(errMd.Pretty())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		rootContent[name] = d.Content
	}
	d.handleGeneric(w, r, "simple", d)
}

func handleGenericLogsDisplay(w http.ResponseWriter, r *http.Request) *homeData {
	rawPage := r.URL.Query().Get("page")
	page := 1
	if rawPage != "" {
		var err error
		page, err = strconv.Atoi(rawPage)
		if err != nil || page < 1 {
			slog.Warn("invalid page number", "rawPage", rawPage)
			w.WriteHeader(http.StatusBadRequest)
			return nil
		}
	}
	d := new(homeData)
	d.data = new(data)
	if sortedLogs == nil {
		sortLogs()
	}
	d.CurrentPage = page
	d.PagesNumber = len(sortedLogs)/maxLogsPerPage + 1
	if d.PagesNumber < page {
		notFound(w, r)
		return nil
	}
	d.Logs = sortedLogs[(page-1)*maxLogsPerPage : min(page*maxLogsPerPage, len(sortedLogs))]
	return d
}
