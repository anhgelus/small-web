package backend

import (
	"fmt"
	"html/template"
	"iter"
	"log/slog"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

var (
	sections = map[string]map[string]*sectionData{}
)

type Section struct {
	Name        string         `toml:"name"`
	Folder      string         `toml:"folder"`
	Description string         `toml:"description"`
	URI         string         `toml:"uri"`
	Data        []*sectionData `toml:"-"`
	Paginate    bool           `toml:"-"`
	PagesNumber int            `toml:"-"`
	CurrentPage int            `toml:"-"`
}

type sectionData struct {
	*data
	EntryInfo
	DataTitle string
	Content   template.HTML
	Slug      string
}

func (d *sectionData) SetData(dt *data) {
	d.data = dt
}

func (d *sectionData) PubDate() string {
	return d.PubLocalDate.String()
}

func (d *sectionData) PubDateRSS() string {
	return d.PubLocalDate.AsTime(time.Local).Format(time.RFC1123Z) // because RFC822 in go isn't RFC822???
}

func (d *sectionData) Title() string {
	return d.data.Title()
}

type image struct {
	Src    string `toml:"src"`
	Alt    string `toml:"alt"`
	Legend string `toml:"legend"`
}

func (s *Section) Load(_ *Config) bool {
	dir, err := os.ReadDir(s.Folder)
	logger := slog.With("folder", s.Folder)
	if err != nil {
		if !os.IsNotExist(err) {
			logger.Error("reading directory", "error", err)
			return false
		}
		logger.Info("log directory does not exist, creating...")
		err = os.MkdirAll(s.Folder, 0774)
		if err != nil {
			slog.Error("creating directory", "error", err)
		}
		return false
	}
	logger.Info("checking directory...", "path", s.Folder)
	err = s.readDir(s.Folder, dir)
	if err != nil {
		slog.Error("reading directory", "error", err)
		return false
	}
	logger.Info("all data loaded")
	return true
}

func (s *Section) readDir(path string, dir []os.DirEntry) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, d := range dir {
		p := filepath.Join(path, d.Name())
		if d.IsDir() {
			dd, err := os.ReadDir(p)
			if err != nil {
				return err
			}
			if err = s.readDir(p, dd); err != nil {
				return err
			}
		} else {
			if !strings.HasSuffix(d.Name(), ".md") {
				return fmt.Errorf("file %s is not a markdown file", d.Name())
			}
			slug := strings.TrimSuffix(p, ".md")
			sec, ok := sections[s.Name]
			if !ok {
				sec = make(map[string]*sectionData, 2)
				sections[s.Name] = sec
			}
			_, ok = sec[slug]
			if ok {
				return fmt.Errorf("data already exists: %s", d.Name())
			}
			dd := new(sectionData)
			dd.data = new(data)

			wg.Add(1)
			go func(p string, d os.DirEntry) {
				defer wg.Done()
				ok = s.parse(dd, &mu, slug, strings.TrimSuffix(d.Name(), ".md"))
				if ok {
					slog.Debug("data parsed", "path", p)
				} else {
					slog.Debug("data skipped", "path", p)
				}
			}(p, d)
		}
	}
	wg.Wait()
	s.sort()
	return nil
}

func (s *Section) Handle(r *chi.Mux) {
	base := "/" + s.URI
	r.Get(base, s.handleList)
	r.Route(base, func(r chi.Router) {
		r.Get("/", s.handleList)

		r.Get("/rss", s.handleRSS)
		r.Get("/rss/", s.handleRSS)

		r.Get("/{slug:[a-zA-Z0-9-]+}", s.handleOne)
		r.Get("/{slug:[a-zA-Z0-9-]+}/", s.handleOne)
	})
}

func (s *Section) handleList(w http.ResponseWriter, r *http.Request) {
	p := s.handlePagination(w, r, 5)
	if p == nil {
		return
	}
	d := new(homeData)
	d.data = new(data)
	d.title = s.Name
	sec := *s
	sec.Data = sec.Data[p.Start:p.End]
	sec.Paginate = true
	sec.CurrentPage = p.Current
	sec.PagesNumber = p.Max
	d.Sections = append(d.Sections, &sec)
	d.PageDescription = sec.Description
	d.handleGeneric(w, r, "home_section", d)
}

func (s *Section) handleRSS(w http.ResponseWriter, r *http.Request) {
	d := handleGenericSectionDisplay(w, r, 5)
	if d == nil {
		return
	}
	d.title = s.Name
	d.handleRSS(w, r, d)
}

func (s *Section) handleOne(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	path := filepath.Join(s.Folder, slug)
	sec, ok := sections[s.Name]
	var d *sectionData
	if ok {
		d, ok = sec[path]
	}
	if !ok {
		d = new(sectionData)
		d.data = new(data)
		if ok = s.parse(d, new(sync.Mutex), path, slug); !ok {
			notFound(w, r)
			return
		}
	}
	d.handleGeneric(w, r, "data", d)
}

func (s *Section) parse(d *sectionData, mu *sync.Mutex, path, slug string) bool {
	d.Article = true
	d.DataTitle = slug
	d.Slug = slug
	b, err := os.ReadFile(path + ".md")
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		panic(err)
	}
	var ok bool
	d.Content, ok = parse(b, &d.EntryInfo, d.data)
	if !ok {
		return false
	}
	d.DataTitle = d.EntryInfo.Title
	mu.Lock()
	sec, ok := sections[s.Name]
	if !ok {
		sec = make(map[string]*sectionData, 2)
		sections[s.Name] = sec
	}
	sec[path] = d
	mu.Unlock()
	return true
}

func (s *Section) sort() {
	s.Data = sort(maps.Values(sections[s.Name]))
}

func sort(values iter.Seq[*sectionData]) []*sectionData {
	return slices.SortedFunc(values, func(l *sectionData, l2 *sectionData) int {
		lt := l.PubLocalDate.AsTime(time.UTC)
		l2t := l2.PubLocalDate.AsTime(time.UTC)
		// we want it reversed
		if lt.Before(l2t) {
			return 1
		} else if lt.After(l2t) {
			return -1
		}
		return 0
	})
}

type pagination struct {
	Current int
	Max     int
	Start   int
	End     int
}

func (s *Section) handlePagination(w http.ResponseWriter, r *http.Request, maxLogsPerPage int) *pagination {
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
	if len(s.Data) == 0 {
		s.sort()
	}
	p := new(pagination)
	p.Current = page
	p.Max = max(1, (len(s.Data)-1)/maxLogsPerPage+1)
	if p.Max < page {
		notFound(w, r)
		return nil
	}
	p.Start = (page - 1) * maxLogsPerPage
	p.End = min(page*maxLogsPerPage, len(s.Data))
	return p
}
