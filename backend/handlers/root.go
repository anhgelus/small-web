package handlers

import (
	"net/http"
	"os"
	"path"

	"anhgelus.world/small-web/backend"
)

func Home() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg := backend.ContextConfig(r.Context())
		err := render(r.Context(), w, "home", Data{
			Custom:          cfg.Sections,
			PageDescription: cfg.Description,
		})
		if err != nil {
			panic(err)
		}
	})
}

func Root() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg := backend.ContextConfig(r.Context())
		art, err := backend.Parse(
			path.Join(cfg.DataFolder, r.PathValue("any")+".md"),
		)
		if err != nil {
			if os.IsNotExist(err) {
				NotFound().ServeHTTP(w, r)
				return
			}
			panic(err)
		}
		err = render(r.Context(), w, "simple", Data{
			Custom:    art.Content(),
			PageTitle: art.Title,
		})
		if err != nil {
			panic(err)
		}
	})
}

func RSS() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg := backend.ContextConfig(r.Context())
		items := make([]*backend.Article, 0, len(cfg.Sections)*7)
		for _, sec := range cfg.Sections {
			items = append(items, sec.FirstN(5)...)
		}
		err := renderRSS(r.Context(), w, RSSData{
			Title:       cfg.Name,
			Description: cfg.Description,
			Items:       items,
		})
		if err != nil {
			panic(err)
		}
	})
}
