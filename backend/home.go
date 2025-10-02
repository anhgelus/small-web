package backend

import (
	"maps"
	"net/http"
	"slices"
	"time"

	"github.com/go-chi/chi/v5"
)

var sortedLogs []*logData

type homeData struct {
	*data
	Logs []*logData
}

func (h *homeData) SetData(d *data) {
	h.data = d
}

func HandleHome(r *chi.Mux) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		d := new(homeData)
		d.data = new(data)
		if sortedLogs == nil {
			sortLogs()
		}
		d.Logs = sortedLogs
		d.handleGeneric(w, r, "home", d)
	})
}

func sortLogs() {
	slices.SortedFunc(maps.Values(logs), func(l *logData, l2 *logData) int {
		lt := l.pubDate.AsTime(time.UTC)
		l2t := l2.pubDate.AsTime(time.UTC)
		// we want it reversed
		if lt.Before(l2t) {
			return 1
		} else if lt.After(l2t) {
			return -1
		}
		return 0
	})
}
