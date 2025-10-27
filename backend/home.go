package backend

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

var (
	//sortedSections = map[string][]*sectionData{}
	rootContent = map[string]*rootData{}
)

type homeData struct {
	*data
	Sections []*Section
}

func (h *homeData) SetData(d *data) {
	h.data = d
}

func HandleHome(r *chi.Mux) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		d := handleGenericSectionDisplay(w, r, 3)
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
		*d = *c
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
		d.Content, ok = parse(b, new(EntryInfo), d.data)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		rootContent[name] = d
	}
	d.handleGeneric(w, r, "simple", d)
}

func handleGenericSectionDisplay(_ http.ResponseWriter, r *http.Request, maxLogsPerPage int) *homeData {
	d := new(homeData)
	d.data = new(data)
	cfg := r.Context().Value(configKey).(*Config)
	for _, sec := range cfg.Sections {
		if len(sec.Data) == 0 {
			sec.sort()
		}
		sec.Data = sec.Data[:min(maxLogsPerPage, len(sec.Data))]
		d.Sections = append(d.Sections, &sec)
	}
	return d
}
