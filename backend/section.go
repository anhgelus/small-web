package backend

import (
	"bytes"
	"html/template"
	"os"
	"path"
	"strings"
	"time"

	"anhgelus.world/small-web/markdown"
	"github.com/nyttikord/avl"
	"github.com/pelletier/go-toml/v2"
)

type Section struct {
	Name        string `toml:"name"`
	TitleName   string `toml:"title_name"`
	Folder      string `toml:"folder"`
	Description string `toml:"description"`
	URI         string `toml:"uri"`
	articles    *avl.KeyAVL[toml.LocalDate, *Article]
	slugToDate  map[string]toml.LocalDate
}

func (s *Section) Get(slug string) *Article {
	k, ok := s.slugToDate[slug]
	if !ok {
		return nil
	}
	v := s.articles.Get(k)
	if v == nil {
		return nil
	}
	return *v
}

func (s *Section) Add(slug string, art *Article) {
	s.articles.Insert(art.PubLocalDate, art)
	if s.slugToDate == nil {
		s.slugToDate = make(map[string]toml.LocalDate)
	}
	s.slugToDate[slug] = art.PubLocalDate
}

func (s *Section) FirstN(n int) []*Article {
	arts := s.Articles()
	return arts[:min(n, len(arts))]
}

func (s *Section) Articles() []*Article {
	return s.articles.Sort()
}

func (s *Section) Init(basePath string) error {
	if s.articles == nil {
		s.articles = avl.NewKey[toml.LocalDate, *Article](func(a, b toml.LocalDate) int {
			return -a.AsTime(time.Local).Compare(b.AsTime(time.Local))
		})
	}
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".bk") {
			continue
		}
		p := path.Join(basePath, entry.Name())
		if entry.IsDir() {
			err = s.Init(p)
			if err != nil {
				return err
			}
			continue
		}
		art, err := Parse(p)
		if err != nil {
			return err
		}
		slug := strings.TrimSuffix(entry.Name(), ".md")
		art.URI = "/" + s.URI + "/" + slug
		s.Add(slug, art)
	}
	return nil
}

type ImageHeader struct {
	Src    string `toml:"src"`
	Alt    string `toml:"alt"`
	Legend string `toml:"legend"`
}

type ArticleContributor struct {
	DID  string `toml:"did"`
	Role string `toml:"role"`
}

type Article struct {
	Title        string                        `toml:"title"`
	Description  string                        `toml:"description"`
	Image        ImageHeader                   `toml:"image"`
	Tags         []string                      `toml:"tags"`
	PubLocalDate toml.LocalDate                `toml:"publication_date"`
	Poem         bool                          `toml:"poem"`
	Contributors map[string]ArticleContributor `toml:"contributors"`
	filePath     string
	URI          string `toml:"-"`
}

func (a *Article) Content() template.HTML {
	b, err := os.ReadFile(a.filePath)
	if err != nil {
		panic(err)
	}
	_, n, ok := bytes.Cut(b, []byte("---"))
	if ok {
		b = n
	}
	res, mdErr := markdown.ParseBytes(b, &markdown.Option{Poem: a.Poem})
	if mdErr != nil {
		println(mdErr.Pretty())
		panic("cannot parse markdown (see logs)")
	}
	return res
}

var now = time.Now()

func (a *Article) PubDateRSS() string {
	t := a.PubLocalDate.AsTime(time.Local)
	// if same day, assume that it's published now
	if t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day() {
		t = now
	}
	return t.Format(time.RFC1123Z) // because RFC822 in go isn't RFC822???
}
