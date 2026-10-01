package handlers

import (
	"net/http"
	"path"

	"anhgelus.world/small-web/backend"
)

type SectionData struct {
	*backend.Section
	Articles []*backend.Article
}

func SectionArticle(sec *backend.Section) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		art := sec.Get(slug)
		if art == nil {
			NotFound().ServeHTTP(w, r)
			return
		}
		err := render(r.Context(), w, "simple", Data{
			Title:           art.Title + " - " + sec.TitleName + " entry",
			Custom:          art,
			PubDate:         art.PubLocalDate.String(),
			Image:           art.Image,
			URL:             "/" + path.Join(sec.URI, r.RequestURI),
			PageDescription: art.Description,
			First:           r.Header.Get("Referer") == "https://"+backend.ContextConfig(r.Context()).Domain,
		})
		if err != nil {
			panic(err)
		}
	})
}

func SectionRSS(sec *backend.Section) http.Handler {
	items := sec.FirstN(7)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := renderRSS(r.Context(), w, RSSData{
			Title:       sec.Name,
			Description: sec.Description,
			URI:         sec.URI,
			Items:       items,
		})
		if err != nil {
			panic(err)
		}
	})
}
