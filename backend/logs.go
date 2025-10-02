package backend

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type logData struct {
	*data
	LogTitle    string
	Description string
	Img         image
}

func (d *logData) SetData(dt *data) {
	d.data = dt
}

type image struct {
	Src    string
	Alt    string
	Legend string
}

func HandleLogs(r *chi.Mux) {
	r.Route("/logs", func(r chi.Router) {
		r.Get("/", handleLogList)
		r.Get("/{slug:[a-zA-Z0-9-]+}", handleLog)
	})
}

func handleLogList(w http.ResponseWriter, r *http.Request) {

}

func handleLog(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	d := new(logData)
	d.data = new(data)
	d.Article = true
	d.LogTitle = slug
	d.title = slug
	d.handleGeneric(w, r, "log", d)
}
