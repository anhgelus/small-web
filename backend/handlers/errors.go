package handlers

import "net/http"

func NotFound() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := render(r.Context(), w, "404", Data{Title: "404", URL: "404", First: false})
		if err != nil {
			panic(err)
		}
		w.WriteHeader(http.StatusNotFound)
	})
}
