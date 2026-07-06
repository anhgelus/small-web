package backend

import (
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"net/http"

	"anhgelus.world/ljus"
)

func ContextMiddleware(cfg *Config, debug bool, db *sql.DB) ljus.Middleware {
	return func(next ljus.Handler, w *ljus.StatusWriter, r *http.Request) {
		ctx := SetContext(
			r.Context(),
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
			cfg := ContextConfig(ctx)
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
		ctx = SetContextConnected(ctx, ok)
		next(w, r.WithContext(ctx))
	}
}
