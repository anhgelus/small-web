package backend

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

const maxLogsPerPage = 5

var sortedLogs []*logData

type homeData struct {
	*data
	Logs        []*logData
	PagesNumber int
	CurrentPage int
}

func (h *homeData) SetData(d *data) {
	h.data = d
}

func HandleHome(r *chi.Mux) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		d := handleGenericLogsDisplay(w, r)
		if d == nil {
			return
		}
		d.handleGeneric(w, r, "home", d)
	})
}

func handleGenericLogsDisplay(w http.ResponseWriter, r *http.Request) *homeData {
	rawPage := r.URL.Query().Get("page")
	page := 1
	if rawPage != "" {
		var err error
		page, err = strconv.Atoi(rawPage)
		if err != nil || page < 1 {
			slog.Warn("invalid page number", "rawPage", rawPage)
			w.WriteHeader(http.StatusBadRequest)
			return nil
		}
	}
	d := new(homeData)
	d.data = new(data)
	if sortedLogs == nil {
		sortLogs()
	}
	d.CurrentPage = page
	d.PagesNumber = len(sortedLogs)/maxLogsPerPage + 1
	if d.PagesNumber < page {
		http.NotFoundHandler().ServeHTTP(w, r)
		return nil
	}
	d.Logs = sortedLogs[(page-1)*maxLogsPerPage : min(page*maxLogsPerPage, len(sortedLogs))]
	return d
}
