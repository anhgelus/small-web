package backend

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

var logs = map[string]string{}

type logData struct {
	*data
	LogTitle    string
	Description string
	Img         image
	Content     template.HTML
}

func (d *logData) SetData(dt *data) {
	d.data = dt
}

type image struct {
	Src    string
	Alt    string
	Legend string
}

func LoadLogs(cfg *Config) bool {
	dir, err := os.ReadDir(cfg.LogFolder)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Error("reading log directory", "error", err)
			return false
		}
		err = os.MkdirAll(cfg.LogFolder, 0660)
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
			_, ok := logs[d.Name()]
			if ok {
				return fmt.Errorf("log already exists: %s", d.Name())
			}
			logs[strings.TrimSuffix(d.Name(), ".md")] = p
		}
	}
	return nil
}

func HandleLogs(r *chi.Mux) {
	r.Route("/logs", func(r chi.Router) {
		r.Get("/", handleLogList)
		r.Get("/{slug:[a-zA-Z0-9-]+}", handleLog)
	})
}

func handleLogList(w http.ResponseWriter, r *http.Request) {

}

func handleLog(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	_, ok := logs[slug]
	if !ok {
		http.NotFoundHandler().ServeHTTP(w, r)
		return
	}
	d := new(logData)
	d.data = new(data)
	d.Article = true
	d.LogTitle = slug
	d.title = slug
	d.handleGeneric(w, r, "log", d)
}
