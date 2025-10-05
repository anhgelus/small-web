package backend

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
)

const (
	Version     = "0.2.0"
	configKey   = "config"
	isUpdateKey = "is_update"
	assetsFS    = "assets_fs"
)

//go:embed templates
var templates embed.FS

func SetupLogger(debug bool) {
	logFormat := httplog.SchemaECS.Concise(!debug)

	logLevel := slog.LevelInfo
	if debug {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: logFormat.ReplaceAttr,
		Level:       logLevel,
	})).With(
		slog.String("app", "anhgelus/small-web"),
		slog.String("version", Version),
	)

	slog.SetDefault(logger)
}

func NewRouter(debug bool, cfg *Config, assets fs.FS) *chi.Mux {
	r := chi.NewRouter()

	logLevel := slog.LevelWarn
	if debug {
		logLevel = slog.LevelDebug
	}

	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(httplog.RequestLogger(slog.Default(), &httplog.Options{
		Level: logLevel,
		// Set log output to Elastic Common Schema (ECS) format.
		Schema:        httplog.SchemaECS.Concise(!debug),
		RecoverPanics: true,
		Skip: func(req *http.Request, respStatus int) bool {
			return respStatus == http.StatusNotFound || respStatus == http.StatusMethodNotAllowed
		},
		// Optionally, log selected request/response headers explicitly.
		LogRequestHeaders:  []string{"Origin"},
		LogResponseHeaders: []string{},
	}))
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Add("Access-Control-Allow-Origin", fmt.Sprintf("https://%s", cfg.Domain))
			if !debug {
				r.Header.Add("Access-Control-Max-Age", fmt.Sprintf("%d", 24*60*60))
			}
			next.ServeHTTP(w, r)
		})
	})
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), configKey, cfg)
			ctx = context.WithValue(ctx, assetsFS, assets)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			val := false
			if r.Header.Get("HX-Request") == "true" {
				val = true
			}
			ctx := context.WithValue(r.Context(), isUpdateKey, val)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	return r
}

// httpEmbedFS is an implementation of fs.FS, fs.ReadDirFS and fs.ReadFileFS helping to manage embed.FS for http server
type httpEmbedFS struct {
	embed.FS
	prefix string
}

func (h *httpEmbedFS) Open(name string) (fs.File, error) {
	return h.FS.Open(h.prefix + "/" + name)
}

func (h *httpEmbedFS) ReadFile(name string) ([]byte, error) {
	return h.FS.ReadFile(h.prefix + "/" + name)
}

func (h *httpEmbedFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return h.FS.ReadDir(h.prefix + "/" + name)
}

// UsableEmbedFS converts embed.FS into usable fs.FS by Golatt
//
// folder may not finish or start with a slash (/)
func UsableEmbedFS(folder string, em embed.FS) fs.FS {
	return &httpEmbedFS{
		prefix: folder,
		FS:     em,
	}
}

func HandleStaticFiles(r *chi.Mux, path string, root fs.FS) {
	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, req *http.Request) {
		ctx := chi.RouteContext(req.Context())
		pathPrefix := strings.TrimSuffix(ctx.RoutePattern(), "/*")
		if pathPrefix+"/" == req.URL.Path {
			r.NotFoundHandler().ServeHTTP(w, req)
			return
		}
		http.StripPrefix(pathPrefix, http.FileServerFS(root)).ServeHTTP(w, req)
	})
}
