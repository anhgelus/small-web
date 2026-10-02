package handlers

import "net/http"

func NotFound() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := render(r.Context(), r, w, "404", CommonData{
			Ttl: "404",
			Uri: "404",
		})
		if err != nil {
			panic(err)
		}
		w.WriteHeader(http.StatusNotFound)
	})
}
