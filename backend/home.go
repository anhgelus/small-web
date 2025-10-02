package backend

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func HandleHome(r *chi.Mux) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		new(data).handleGeneric(w, r, "home")
	})
}
