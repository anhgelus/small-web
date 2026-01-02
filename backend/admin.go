package backend

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"git.anhgelus.world/anhgelus/small-web/backend/storage"
	"github.com/go-chi/chi/v5"
)

type adminData struct {
	*data
	Visits      []storage.StatsRow
	Rows        []storage.StatsRow
	PagesNumber int
	CurrentPage int
}

type to struct {
	n     int
	since time.Time
}

type tos struct {
	mu  sync.Mutex
	tos map[string]*to
}

var timeouts = tos{tos: make(map[string]*to)}

func rateLimitDuration(n int) time.Duration {
	return time.Duration(math.Pow10(n/4)) * time.Second
}

func isRateLimited(ctx context.Context) bool {
	ip := ctx.Value(storage.IPAddressKey).(string)

	timeouts.mu.Lock()
	defer timeouts.mu.Unlock()

	v, ok := timeouts.tos[ip]
	if !ok {
		return false
	}
	return time.Since(v.since) <= rateLimitDuration(v.n)
}

func rateLimit(ctx context.Context) bool {
	ip := ctx.Value(storage.IPAddressKey).(string)

	timeouts.mu.Lock()
	defer timeouts.mu.Unlock()

	v, ok := timeouts.tos[ip]
	if !ok {
		timeouts.tos[ip] = &to{n: 1}
		return false
	}
	if time.Since(v.since) <= rateLimitDuration(v.n) {
		return true
	}
	v.n++
	if v.n%4 != 0 {
		return false
	}
	v.since = time.Now()
	GetLogger(ctx).Warn("rate limiting IP", "ip", ip, "duration", rateLimitDuration(v.n).String())
	go func(v *to, ip string) {
		time.Sleep(3 * time.Hour)
		v.n = max(v.n-4, 0)
		if v.n == 0 {
			timeouts.mu.Lock()
			defer timeouts.mu.Unlock()
			delete(timeouts.tos, ip)
		}
	}(v, ip)
	return true
}

func resetRateLimit(ctx context.Context) {
	ip := ctx.Value(storage.IPAddressKey).(string)

	timeouts.mu.Lock()
	defer timeouts.mu.Unlock()

	delete(timeouts.tos, ip)
}

func HandleAdmin(r *chi.Mux) {
	r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if !ctx.Value(loginKey).(bool) {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		d := new(adminData)
		d.data = new(data)
		rawPage := r.URL.Query().Get("page")
		page := 1
		var err error
		if rawPage != "" {
			page, err = strconv.Atoi(rawPage)
			if err != nil || page < 1 {
				GetLogger(ctx).Warn("invalid page number", "requested", rawPage)
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
		}
		d.Rows, err = storage.GetStatsRows(ctx, uint(page))
		if err != nil {
			panic(err)
		}
		d.Visits, err = storage.GetUnionStatsRows(ctx)
		if err != nil {
			panic(err)
		}
		d.PagesNumber = page + max(len(d.Rows)-storage.StatsPerPage+1, 0)
		d.CurrentPage = page
		d.handleGeneric(w, r, "admin", d)
	})
}
