package backend

import "github.com/pelletier/go-toml/v2"

type Section struct {
	Name        string `toml:"name"`
	TitleName   string `toml:"title_name"`
	Folder      string `toml:"folder"`
	Description string `toml:"description"`
	URI         string `toml:"uri"`
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
