package handlers

import "net/http"

func NotFound(w http.ResponseWriter, r *http.Request) {
	err := render(r.Context(), w, "404", Data{PageTitle: "404"})
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusNotFound)
}
