package backend

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func HandleHome(r *chi.Mux) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		t := template.Must(template.ParseFS(templates, "templates/home.html", "templates/base.html"))
		err := t.ExecuteTemplate(w, "base.html", &data{})
		if err != nil {
			panic(err)
		}
	})
}
