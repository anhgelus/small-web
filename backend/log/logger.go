package log

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
)

const (
	loggerKey  = "logger"
	statusCode = "status_code"
)

func GetLogger(ctx context.Context) *slog.Logger {
	return ctx.Value(loggerKey).(*slog.Logger)
}

type customWriter struct {
	http.ResponseWriter
	statusCode int
}

func (c *customWriter) WriteHeader(statusCode int) {
	if statusCode == c.statusCode {
		return
	}
	c.statusCode = statusCode
	c.ResponseWriter.WriteHeader(statusCode)
}

func GetStatusCode(ctx context.Context) func() int {
	return ctx.Value(statusCode).(func() int)
}

func newLogger(l *slog.Logger, r *http.Request) *slog.Logger {
	return l.With("uri", r.RequestURI, "method", r.Method)
}

func SetContextLogger(ctx context.Context, l *slog.Logger, r *http.Request) context.Context {
	return context.WithValue(ctx, loggerKey, newLogger(l, r))
}

func SetLogger(l *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := newLogger(l, r)
			ww := &customWriter{ResponseWriter: w, statusCode: http.StatusOK}
			ctx := context.WithValue(r.Context(), statusCode, func() int {
				return ww.statusCode
			})
			ctx = context.WithValue(ctx, loggerKey, logger)
			defer func(logger *slog.Logger) {
				rec := recover()
				if rec == nil {
					return
				}
				if rec == http.ErrAbortHandler {
					panic(rec)
				}
				logger.Error("crashed, recovered", "error", rec, "status", http.StatusInternalServerError)
				log.New(os.Stderr, "", 0).Printf("%s\n", debug.Stack())
				http.Error(ww, "internal error", http.StatusInternalServerError)
			}(logger)

			next.ServeHTTP(ww, r.WithContext(ctx))

			if ww.statusCode == http.StatusNotFound || ww.statusCode == http.StatusTooManyRequests {
				return
			}
			var lvl slog.Level
			switch {
			case ww.statusCode >= 500:
				lvl = slog.LevelError
			case ww.statusCode >= 400:
				lvl = slog.LevelWarn
			case ww.statusCode >= 300:
				return
			default:
				lvl = slog.LevelInfo
			}
			logger.Log(context.Background(), lvl, "handled", "status", ww.statusCode)
		})
	}
}
