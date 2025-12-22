package backend

import (
	"log/slog"
	"net/http"
	"strconv"

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
				slog.Warn("invalid page number", "rawPage", rawPage)
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
