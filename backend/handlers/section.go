package handlers

import (
	"net/http"
	"strconv"

	"anhgelus.world/small-web/backend"
)

type ArticleData struct {
	*backend.Article
	URI  string
	Slug string
}

type SectionData struct {
	*backend.Section
	Articles []ArticleData
	// Pagination
	Paginate    bool
	LenMax      int
	CurrentPage int
	PagesNumber int
}

func paginate(articles []ArticleData, maxLogs int, r *http.Request) (page int, arts []ArticleData) {
	rawPage := r.URL.Query().Get("page")
	if rawPage == "" {
		page = 1
	} else {
		var err error
		page, err = strconv.Atoi(rawPage)
		if err != nil || page < 1 {
			backend.ContextLogger(r.Context()).Warn("invalid page number", "requested", rawPage)
			return
		}
	}
	if max(1, (len(articles)-1)/maxLogs+1) < page {
		return
	}
	arts = articles[(page-1)*maxLogs : min(page*maxLogs, len(articles))]
	return
}

func SectionHome(sec *backend.Section) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, arts := paginate(nil, 7, r)
		if page < 1 {
			http.Error(w, "Bad request: invalid page number", http.StatusBadRequest)
			return
		}
		if arts == nil {
			NotFound(w, r)
			return
		}
		v := SectionData{
			Section:     sec,
			Articles:    arts,
			Paginate:    true,
			LenMax:      7,
			CurrentPage: page,
		}
		err := render(r.Context(), w, "home_section", Data{PageTitle: sec.TitleName, Custom: v})
		if err != nil {
			panic(err)
		}
	})
}

func SectionArticle(sec *backend.Section) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		art, ok := sec.Articles[slug]
		if !ok {
			NotFound(w, r)
			return
		}
		err := render(r.Context(), w, "data", Data{
			PageTitle: art.Title + " - " + sec.TitleName + " entry",
			Custom:    art,
			PubDate:   art.PubLocalDate.String(),
		})
		if err != nil {
			panic(err)
		}
	})
}
