package handlers

import (
	"net/http"
	"strconv"

	"anhgelus.world/small-web/backend"
	"anhgelus.world/small-web/backend/storage"
)

type AdminData struct {
	Visits      []storage.StatsRow
	Rows        []storage.StatsRow
	PagesNumber int
	CurrentPage int
}

func Admin() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if !backend.ContextConnnected(ctx) {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		rawPage := r.URL.Query().Get("page")
		page := 1
		var err error
		if rawPage != "" {
			page, err = strconv.Atoi(rawPage)
			if err != nil || page < 1 {
				backend.ContextLogger(ctx).Warn("invalid page number", "requested", rawPage)
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
		}
		rows, err := storage.GetStatsRows(ctx, uint(page))
		if err != nil {
			panic(err)
		}
		visits, err := storage.GetUnionStatsRows(ctx)
		if err != nil {
			panic(err)
		}
		err = render(ctx, w, "admin", Data{Custom: AdminData{
			Visits:      visits,
			Rows:        rows,
			PagesNumber: page + max(len(rows)-storage.StatsPerPage+1, 0),
			CurrentPage: page,
		}})
		if err != nil {
			panic(err)
		}
	})
}
