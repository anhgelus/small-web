package handlers

import "net/http"

func NotFound() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := render(r.Context(), w, "404", Data{Title: "404"})
		if err != nil {
			panic(err)
		}
		w.WriteHeader(http.StatusNotFound)
	})
}
