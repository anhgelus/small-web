package backend

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"git.anhgelus.world/anhgelus/small-web/backend/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	configKey   = "config"
	assetsFSKey = "assets_fs"
	debugKey    = "debug"
	loginKey    = "login"
)

//go:embed templates
var templates embed.FS

func SetupLogger(debug bool) {
	logLevel := slog.LevelInfo
	if debug {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})).With(
		slog.String("app", "anhgelus/small-web"),
	)

	slog.SetDefault(logger)
}

func NewRouter(debug bool, cfg *Config, db *sql.DB, assets fs.FS) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(SetLogger(slog.Default()))
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
	setContext := func(ctx context.Context, r *http.Request) context.Context {
		ip := r.Header.Get("X-Real-Ip")
		if ip == "" {
			ip = r.Header.Get("X-Forwarded-For")
		}
		if ip == "" {
			ip = r.RemoteAddr
		}
		if strings.Contains(ip, ":") {
			ip = strings.Split(ip, ":")[0]
		}
		ctx = context.WithValue(ctx, storage.IPAddressKey, ip)
		ctx = context.WithValue(ctx, configKey, cfg)
		ctx = context.WithValue(ctx, assetsFSKey, assets)
		ctx = context.WithValue(ctx, debugKey, debug)
		return context.WithValue(ctx, storage.DBKey, db)
	}
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.RequestURI, "/") {
				r.RequestURI = "/" + r.RequestURI
			}
			if !strings.HasPrefix(r.URL.Path, "/") {
				r.URL.Path = "/" + r.URL.Path
			}
			next.ServeHTTP(w, r.WithContext(
				setContext(r.Context(), r),
			))
		})
	})
	// login
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if isRateLimited(ctx) {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}
			_, pass, ok := r.BasicAuth()
			if ok {
				cfg := ctx.Value(configKey).(*Config)
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
			ctx = context.WithValue(ctx, loginKey, ok)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	// stats
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			go func(ctx context.Context, r *http.Request) {
				if strings.HasPrefix(r.RequestURI, "/static") || r.RequestURI == "/robots.txt" {
					return
				}
				logger := GetLogger(ctx)
				debug := ctx.Value(debugKey).(bool)
				if ctx.Value(loginKey).(bool) && !debug {
					logger.Debug("not updating stats because user is admin logged")
					return
				}
				statusCode := GetStatusCode(ctx)()
				if statusCode >= 299 && r.RequestURI != storage.HumanPageLoad {
					logger.Debug("not updating stats for status code above 299", "status", statusCode)
					return
				}
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
				defer cancel()
				err := storage.UpdateStats(setContext(ctx, r), r, cfg.Domain)
				if err != nil {
					logger.Error("updating stats", "error", err)
				}
			}(r.Context(), r)
		})
	})
	// anti dumb attackers bot
	phpUri := regexp.MustCompile(`/.+\.php(/.*)?`)
	dotUri := regexp.MustCompile(`/(.*/)*\..*`)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if phpUri.MatchString(r.RequestURI) || dotUri.MatchString(r.RequestURI) {
				handleSus(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// txt files
	r.Get("/{file:[a-z]+}.txt", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		cfg := ctx.Value(configKey).(*Config)
		logger := GetLogger(ctx)
		logger.Info("requesting txt file", "User-Agent", r.Header.Get("User-Agent"))
		b, err := os.ReadFile(path.Join(cfg.PublicFolder, chi.URLParam(r, "file")+".txt"))
		if os.IsNotExist(err) {
			notFound(w, r)
			return
		} else if err != nil {
			panic(err)
		}
		_, err = w.Write(b)
		if err != nil {
			panic(err)
		}
	})

	return r
}

func handleSus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := GetLogger(ctx)
	logger.Warn("sus request", "User-Agent", r.Header.Get("User-Agent"))
	if rateLimit(ctx) {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}
	notFound(w, r)
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
