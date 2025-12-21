package backend

import (
	"context"
	"database/sql"
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
	Version     = "0.5.0"
	configKey   = "config"
	assetsFSKey = "assets_fs"
	debugKey    = "debug"
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

func NewRouter(debug bool, cfg *Config, db *sql.DB, assets fs.FS) *chi.Mux {
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
	// security headers
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// prevent tracking
			w.Header().Add("Referrer-Policy", "same-origin")
			// prevent iframe
			w.Header().Add("X-Frame-Options", "deny")
			// prevent bad content being parsed
			w.Header().Add("X-Content-Type-Options", "nosniff")
			w.Header().Add("X-Permitted-Cross-Domain-Policies", "none")
			// content security, cors & co
			w.Header().Add("Content-Security-Policy", fmt.Sprintf("default-src 'self' *.%s; object-src 'none';", cfg.Domain))
			w.Header().Add("Access-Control-Allow-Origin", fmt.Sprintf("https://%s", cfg.Domain))
			if !debug {
				w.Header().Add("Access-Control-Max-Age", fmt.Sprintf("%d", 24*60*60))
			}
			next.ServeHTTP(w, r)
		})
	})
	// context
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), configKey, cfg)
			ctx = context.WithValue(ctx, assetsFSKey, assets)
			ctx = context.WithValue(ctx, debugKey, debug)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	// stats
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := UpdateStats(r.Context(), db, r); err != nil {
				slog.Error("updating stats", "error", err)
			}
			next.ServeHTTP(w, r)
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
