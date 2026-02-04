package backend

import (
	"html/template"
	"iter"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/go-chi/chi/v5"
)

var (
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
		cfg := r.Context().Value(configKey).(*Config)
		d := handleGenericSectionDisplay(w, r, cfg.Sections, 4)
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
	r.Get("/rss", handleGenericRSS)
	r.Get("/rss/", handleGenericRSS)
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
		d.URL = "/" + name
		d.Content, ok = parse(b, new(EntryInfo), d.data)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		rootContent[name] = d
	}
	d.handleGeneric(w, r, "simple", d)
}

func handleGenericRSS(w http.ResponseWriter, r *http.Request) {
	cfg := r.Context().Value(configKey).(*Config)
	var data iter.Seq[*sectionData]
	for _, sec := range cfg.Sections {
		if len(sec.Data) == 0 {
			sec.sort()
		}
		var sl []*sectionData
		for _, d := range sec.Data[:min(3, len(sec.Data))] {
			dd := *d
			dd.Slug = sec.URI + "/" + dd.Slug
			sl = append(sl, &dd)
		}
		if data == nil {
			data = slices.Values(sl)
		} else {
			data = slices.Values(slices.AppendSeq(sl, data))
		}
	}
	var s Section
	s.Data = sort(data)
	s.Name = cfg.Name
	s.Description = cfg.Description
	s.URI = ""
	s.handleRSS(w, r)
}

func handleGenericSectionDisplay(_ http.ResponseWriter, _ *http.Request, sections []Section, maxLogsPerPage int) *homeData {
	d := new(homeData)
	d.data = new(data)
	for _, sec := range sections {
		if len(sec.Data) == 0 {
			sec.sort()
		}
		sec.LenMax = maxLogsPerPage
		sec.Data = sec.Data[:min(maxLogsPerPage, len(sec.Data))]
		d.Sections = append(d.Sections, &sec)
	}
	return d
}
