package backend

import (
	"html/template"
	"os"
	"path"
	"strings"
	"time"

	"anhgelus.world/small-web/markdown"
	"github.com/pelletier/go-toml/v2"
)

type Section struct {
	Name        string `toml:"name"`
	TitleName   string `toml:"title_name"`
	Folder      string `toml:"folder"`
	Description string `toml:"description"`
	URI         string `toml:"uri"`
	Articles    map[string]*Article
}

func (s *Section) Init(basePath string) error {
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
		s.Articles[slug] = art
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
}

func (a *Article) Content() template.HTML {
	b, err := os.ReadFile(a.filePath)
	if err != nil {
		panic(err)
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
