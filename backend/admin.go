package backend

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"strings"
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

func handleTimeout(ctx context.Context) bool {
	ip := ctx.Value(ipAdressKey).(string)
	parsed := strings.Split(ip, ":")
	ip = parsed[0]

	timeouts.mu.Lock()
	defer timeouts.mu.Unlock()

	v, ok := timeouts.tos[ip]
	if !ok {
		timeouts.tos[ip] = &to{n: 1}
		return false
	}
	dur := func() time.Duration { return time.Duration(math.Pow10(v.n/4)) * time.Second }
	if time.Since(v.since) <= dur() {
		return true
	}
	v.n++
	if v.n%4 != 0 {
		return false
	}
	v.since = time.Now()
	GetLogger(ctx).Warn("rate limiting IP", "ip", ip, "duration", dur().String())
	go func(v *to) {
		time.Sleep(3 * time.Hour)
		v.n = max(v.n-4, 0)
	}(v)
	return true
}

func resetTimeout(ctx context.Context) {
	ip := ctx.Value(ipAdressKey).(string)
	parsed := strings.Split(ip, ":")
	ip = parsed[0]

	timeouts.mu.Lock()
	defer timeouts.mu.Unlock()

	v, ok := timeouts.tos[ip]
	if !ok {
		return
	}
	v.n = 0
	v.since = time.Unix(0, 0)
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
