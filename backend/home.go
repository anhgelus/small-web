package backend

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func HandleHome(r *chi.Mux) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		d := &data{
			title:       "",
			Article:     false,
			Domain:      "anhgelus.world",
			URL:         "/",
			Image:       "",
			Description: "",
		}
		d.handleGeneric(w, r, "home")
	})
}
