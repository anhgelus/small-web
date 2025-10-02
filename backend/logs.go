package backend

import (
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"git.anhgelus.world/anhgelus/small-world/markdown"
	"github.com/go-chi/chi/v5"
	"github.com/pelletier/go-toml/v2"
)

var (
	logs       = map[string]string{}
	loadedLogs = map[string]*logData{}
)

type logData struct {
	*data
	LogTitle    string        `toml:"title"`
	Description string        `toml:"description"`
	Img         image         `toml:"image"`
	Content     template.HTML `toml:"-"`
}

func (d *logData) SetData(dt *data) {
	d.data = dt
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
	path, ok := logs[slug]
	if !ok {
		http.NotFoundHandler().ServeHTTP(w, r)
		return
	}
	var d *logData
	d, ok = loadedLogs[slug]
	if !ok {
		d = new(logData)
		d.data = new(data)
		d.Article = true
		d.LogTitle = slug
		d.title = slug
		if ok = parseLog(d, path); !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		loadedLogs[slug] = d
	}
	d.handleGeneric(w, r, "log", d)
}

func parseLog(d *logData, path string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var dd string
	splits := strings.SplitN(string(b), "---", 2)
	if len(splits) == 2 {
		dd = splits[1]
		err = toml.Unmarshal([]byte(splits[0]), &d)
		if err != nil {
			panic(err)
		}
		d.title = d.LogTitle
		d.Image = d.Img.Src
	} else {
		dd = string(b)
	}
	d.Content, err = markdown.Parse(dd)
	var errMd *markdown.ParseError
	errors.As(err, &errMd)
	if errMd != nil {
		slog.Error("parsing markdown")
		fmt.Println(errMd.Pretty())
		return false
	}
	return true
}
