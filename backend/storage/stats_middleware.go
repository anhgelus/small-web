package storage

import (
	"context"
	"net/http"
	"strings"
	"time"

	"anhgelus.world/ljus"
	"anhgelus.world/small-web/backend"
)

func StatsMiddleware() ljus.Middleware {
	return func(next ljus.Handler, w *ljus.StatusWriter, r *http.Request) {
		next(w, r)
		go func(r *http.Request) {
			if strings.HasPrefix(r.RequestURI, "/static") || r.RequestURI == "/robots.txt" {
				return
			}
			ctx := r.Context()
			logger := backend.ContextLogger(ctx)
			debug := backend.ContextDebug(ctx)
			if backend.ContextConnnected(ctx) && !debug {
				logger.Debug("not updating stats because user is admin logged")
				return
			}
			statusCode := w.Code
			if statusCode >= 299 && r.RequestURI != HumanPageLoad {
				logger.Debug("not updating stats for status code above 299", "status", statusCode)
				return
			}
			cfg := backend.ContextConfig(ctx)
			ctx2, cancel := context.WithTimeout(
				context.Background(),
				1*time.Second)
			defer cancel()
			err := UpdateStats(
				backend.CloneContext(ctx2, r.Context()),
				r,
				cfg.Domain)
			if err != nil {
				logger.Error("updating stats", "error", err)
			}
		}(r)
	}
}
