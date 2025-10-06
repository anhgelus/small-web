package backend

import (
	"fmt"
	"html/template"
	"log/slog"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

var (
	logs = map[string]*logData{}
)

type logData struct {
	*data
	EntryInfo
	LogTitle string
	Content  template.HTML
	Slug     string
}

func (d *logData) SetData(dt *data) {
	d.data = dt
}

func (d *logData) PubDate() string {
	return d.PubLocalDate.String()
}

func (d *logData) PubDateRSS() string {
	return d.PubLocalDate.AsTime(time.Local).Format(time.RFC1123Z) // because RFC822 in go isn't RFC822???
}

func (d *logData) Title() string {
	return d.data.Title()
}

type image struct {
	Src    string `toml:"src"`
	Alt    string `toml:"alt"`
	Legend string `toml:"legend"`
}

func LoadLogs(cfg *Config) bool {
	dir, err := os.ReadDir(cfg.LogFolder)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Error("reading log directory", "error", err)
			return false
		}
		slog.Info("log directory does not exist, creating...")
		err = os.MkdirAll(cfg.LogFolder, 0774)
		if err != nil {
			slog.Error("creating log directory", "error", err)
		}
		return false
	}
	slog.Info("checking log directory...", "path", cfg.LogFolder)
	err = readLogDir(cfg.LogFolder, dir)
	if err != nil {
		slog.Error("reading log directory", "error", err, "path", cfg.LogFolder)
		return false
	}
	slog.Info("all logs loaded")
	return true
}

func readLogDir(path string, dir []os.DirEntry) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, d := range dir {
		p := filepath.Join(path, d.Name())
		if d.IsDir() {
			dd, err := os.ReadDir(p)
			if err != nil {
				return err
			}
			if err = readLogDir(p, dd); err != nil {
				return err
			}
		} else {
			if !strings.HasSuffix(d.Name(), ".md") {
				return fmt.Errorf("file %s is not a markdown file", d.Name())
			}
			slug := strings.TrimSuffix(p, ".md")
			_, ok := logs[slug]
			if ok {
				return fmt.Errorf("log already exists: %s", d.Name())
			}
			dd := new(logData)
			dd.data = new(data)

			wg.Add(1)
			go func(p string, d os.DirEntry) {
				defer wg.Done()
				ok = parseLog(dd, &mu, slug, strings.TrimSuffix(d.Name(), ".md"))
				if ok {
					slog.Debug("log parsed", "path", p)
				} else {
					slog.Debug("log skipped", "path", p)
				}
			}(p, d)
		}
	}
	wg.Wait()
	sortLogs()
	return nil
}

func HandleLogs(r *chi.Mux) {
	r.Get("/logs", handleLogList)
	r.Route("/logs", func(r chi.Router) {
		r.Get("/", handleLogList)

		r.Get("/rss", handleLogRSS)
		r.Get("/rss/", handleLogRSS)

		r.Get("/{slug:[a-zA-Z0-9-]+}", handleLog)
		r.Get("/{slug:[a-zA-Z0-9-]+}/", handleLog)
	})
}

func handleLogList(w http.ResponseWriter, r *http.Request) {
	d := handleGenericLogsDisplay(w, r, 5)
	if d == nil {
		return
	}
	d.title = "logs"
	d.handleGeneric(w, r, "home_log", d)
}

func handleLogRSS(w http.ResponseWriter, r *http.Request) {
	d := handleGenericLogsDisplay(w, r, 5)
	if d == nil {
		return
	}
	d.title = "logs"
	d.handleRSS(w, r, d)
}

func handleLog(w http.ResponseWriter, r *http.Request) {
	cfg := r.Context().Value(configKey).(*Config)
	slug := chi.URLParam(r, "slug")
	path := filepath.Join(cfg.LogFolder, slug)
	d, ok := logs[path]
	if !ok {
		d = new(logData)
		d.data = new(data)
		if ok = parseLog(d, new(sync.Mutex), path, slug); !ok {
			notFound(w, r)
			return
		}
	}
	d.handleGeneric(w, r, "log", d)
}

func parseLog(d *logData, mu *sync.Mutex, path, slug string) bool {
	d.Article = true
	d.title = slug
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
	d.LogTitle = d.EntryInfo.Title
	mu.Lock()
	logs[path] = d
	mu.Unlock()
	return true
}

func sortLogs() {
	sortedLogs = slices.SortedFunc(maps.Values(logs), func(l *logData, l2 *logData) int {
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
