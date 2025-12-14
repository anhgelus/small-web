package backend

import (
	"html/template"
	"log/slog"
	"os"

	"git.anhgelus.world/anhgelus/small-web/markdown"
	"github.com/pelletier/go-toml/v2"
)

type Link struct {
	Name string `toml:"name"`
	URL  string `toml:"url"`
}

func (l *Link) Render(url string) template.HTML {
	return renderLink(l.Name, l.URL, url)
}

type Logo struct {
	Header  string `toml:"header"`
	Favicon string `toml:"favicon"`
}

type Replacer struct {
	Symbol  string `toml:"symbol"`
	Replace string `tomle:"replace"`
}

type Config struct {
	Domain       string   `toml:"domain"`
	Name         string   `toml:"name"`
	Description  string   `toml:"description"`
	DefaultImage string   `toml:"default_image"`
	Quotes       []string `toml:"quotes"`
	Language     string   `toml:"language"`

	Sections []Section `toml:"section"`

	RootFolder   string `toml:"root_folder"`
	PublicFolder string `toml:"public_folder"`

	Links []Link `toml:"links"`
	Logo  Logo   `toml:"logo"`

	Replacers []Replacer `toml:"replacers"`
}

func (c *Config) DefaultValues() {
	c.Domain = "example.org"
	c.Name = "example"
	c.Description = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magnam aliquam quaerat voluptatem. Ut enim aeque doleamus animo, cum corpore dolemus, fieri tamen permagna accessio potest, si aliquod aeternum et infinitum impendere malum nobis opinemur. Quod idem licet transferre in voluptatem, ut."
	c.Links = []Link{
		{
			Name: "Home",
			URL:  "/",
		},
		{
			Name: "Logs",
			URL:  "/log/",
		},
	}
	c.Logo = Logo{
		Header:  "logo.jpg",
		Favicon: "favicon.jpg",
	}
	c.Sections = []Section{{
		Name:        "logs",
		Description: "Aut maxime voluptatibus ut dicta voluptates et ut alias. Sunt et incidunt similique et doloremque nostrum fugit autem. Ut omnis quo nisi. Accusantium voluptas fugit autem maiores numquam doloribus.",
		Folder:      "data/logs",
		URI:         "logs",
	}}
	c.RootFolder = "data"
	c.PublicFolder = "public"
	c.Quotes = []string{"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do."}
	c.Replacers = []Replacer{{"~", "&thinsp;"}}
}

var defaultMarkdownOption markdown.Option

func LoadConfig(path string) (*Config, bool) {
	b, err := os.ReadFile(path)
	var config Config
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Error("reading config file", "error", err)
			return nil, false
		}
		slog.Warn("config file not found", "path", path)
		slog.Info("creating a new config file", "path", path)
		config.DefaultValues()
		b, err = toml.Marshal(&config)
		if err != nil {
			slog.Error("marshalling config file", "error", err)
			return nil, false
		}
		err = os.WriteFile(path, b, 0660)
		if err != nil {
			slog.Error("writing config file", "error", err, "path", path)
		} else {
			slog.Info("config file created", "path", path)
		}
		return nil, false
	}
	err = toml.Unmarshal(b, &config)
	if err != nil {
		slog.Error("unmarshalling config file", "error", err)
		return nil, false
	}
	defaultMarkdownOption.ImageSource = getStatic
	defaultMarkdownOption.Replaces = make(map[rune]string, len(config.Replacers))
	for _, r := range config.Replacers {
		if len(r.Symbol) != 1 {
			slog.Error("invalid symbol in config", "symbol", r.Symbol)
			return nil, false
		}
		defaultMarkdownOption.Replaces[[]rune(r.Symbol)[0]] = r.Replace
	}
	return &config, true
}
