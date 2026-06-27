package backend

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"anhgelus.world/ljus"
	"git.anhgelus.world/anhgelus/small-web/backend/common"
	"git.anhgelus.world/anhgelus/small-web/backend/storage"
)

//go:embed templates
var templates embed.FS

func ContextMiddleware(cfg *Config, debug bool, db *sql.DB) ljus.Middleware {
	return func(next ljus.Handler, w *ljus.StatusWriter, r *http.Request) {
		ctx := common.SetContext(
			r.Context(),
			r,
			cfg,
			assets,
			debug,
			db)
		next(w, r.WithContext(ctx))
	}
}

func RateLimitMiddleware() ljus.Middleware {
	return func(next ljus.Handler, w *ljus.StatusWriter, r *http.Request) {
		ctx := r.Context()
		if isRateLimited(ctx) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		_, pass, ok := r.BasicAuth()
		if ok {
			cfg := common.ContextConfig[*Config](ctx)
			passHash := sha256.Sum256([]byte(pass))
			rightPassHash := sha256.Sum256([]byte(cfg.AdminPassword))
			ok = subtle.ConstantTimeCompare(passHash[:], rightPassHash[:]) == 1
			if ok {
				resetRateLimit(ctx)
			} else if rateLimit(ctx) {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}
		}
		ctx = common.SetContextConnected(ctx, ok)
		next(w, r.WithContext(ctx))
	}
}

func StatsMiddleware() ljus.Middleware {
	return func(next ljus.Handler, w *ljus.StatusWriter, r *http.Request) {
		next(w, r)
		go func(r *http.Request) {
			if strings.HasPrefix(r.RequestURI, "/static") || r.RequestURI == "/robots.txt" {
				return
			}
			ctx := r.Context()
			logger := common.ContextLogger(ctx)
			debug := common.ContextDebug(ctx)
			if common.ContextConnnected(ctx) && !debug {
				logger.Debug("not updating stats because user is admin logged")
				return
			}
			statusCode := w.Code
			if statusCode >= 299 && r.RequestURI != storage.HumanPageLoad {
				logger.Debug("not updating stats for status code above 299", "status", statusCode)
				return
			}
			cfg := common.ContextConfig[*Config](ctx)
			ctx2, cancel := context.WithTimeout(
				context.Background(),
				1*time.Second)
			defer cancel()
			err := storage.UpdateStats(
				common.CloneContext(ctx2, r.Context()),
				r,
				cfg.Domain)
			if err != nil {
				logger.Error("updating stats", "error", err)
			}
		}(r)
	}
}

func TxtFilesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cfg := common.ContextConfig[*Config](ctx)
	logger := common.ContextLogger(ctx)
	logger.Info("requesting txt file", "User-Agent", r.Header.Get("User-Agent"))
	b, err := os.ReadFile(path.Join(cfg.PublicFolder, r.PathValue("any")+".txt"))
	if os.IsNotExist(err) {
		NotFoundHandler(w, r)
		return
	} else if err != nil {
		panic(err)
	}
	_, err = w.Write(b)
	if err != nil {
		panic(err)
	}
}

func handleSus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := common.ContextLogger(ctx)
	logger.Debug("sus request", "User-Agent", r.Header.Get("User-Agent"))
	if rateLimit(ctx) {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}
	NotFoundHandler(w, r)
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

func StaticFilesHandler(path string, root fs.FS) ljus.Route {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	return ljus.NewRouteFunc(path+"{file...}", func(w http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.RequestURI, "/") {
			NotFoundHandler(w, req)
			return
		}
		http.StripPrefix(path, http.FileServerFS(root)).ServeHTTP(w, req)
	}).SetName("static files " + path)
}
