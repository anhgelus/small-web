package backend

import (
	"html/template"
	"log/slog"
	"os"
	"strings"

	"anhgelus.world/small-web/dom"
	"anhgelus.world/small-web/markdown"
	"anhgelus.world/xrpc/atproto"
	"github.com/pelletier/go-toml/v2"
)

type Link struct {
	Name string `toml:"name"`
	URL  string `toml:"url"`
}

func (l *Link) Render(url string) template.HTML {
	return func(content, href string) template.HTML {
		anchor := dom.NewLiteralContentElement("a", template.HTML(content))
		anchor.SetAttribute("href", href)
		if href == url || (href != "/" && url != "/" && strings.HasPrefix(url, href)) {
			anchor.ClassList().Add("target")
		}
		if markdown.ExternalLink.MatchString(href) {
			anchor.SetAttribute("target", "_blank").SetAttribute("rel", "noreferrer")
		}
		return anchor.Render()
	}(l.Name, l.URL)
}

type Logo struct {
	Header  string `toml:"header"`
	Favicon string `toml:"favicon"`
}

type Replacer struct {
	Symbol  string `toml:"symbol"`
	Replace string `tomle:"replace"`
}

type ATProto struct {
	PublicationRKey atproto.RecordKey `toml:"publication_rkey"`
	DID             string            `toml:"did"`
	Password        string            `toml:"password"`
	DisplayName     string            `toml:"display_name"`
}

type Config struct {
	Domain        string   `toml:"domain"`
	Name          string   `toml:"name"`
	Description   string   `toml:"description"`
	Quotes        []string `toml:"quotes"`
	Language      string   `toml:"language"`
	Database      string   `toml:"database"`
	AdminPassword string   `toml:"admin_password"`

	DataFolder   string `toml:"data_folder"`
	PublicFolder string `toml:"public_folder"`

	Logo Logo `toml:"logo"`

	ATProto ATProto `toml:"atproto"`

	Sections []*Section `toml:"section"`

	Links []Link `toml:"links"`

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
	c.Sections = []*Section{{
		Name:        "logs",
		TitleName:   "log",
		Description: "Aut maxime voluptatibus ut dicta voluptates et ut alias. Sunt et incidunt similique et doloremque nostrum fugit autem. Ut omnis quo nisi. Accusantium voluptas fugit autem maiores numquam doloribus.",
		Folder:      "data/logs",
		URI:         "logs",
	}}
	c.DataFolder = "data"
	c.PublicFolder = "public"
	c.Database = "database.sqlite"
	c.AdminPassword = "Ch@ngeM€Please!"
	c.Quotes = []string{"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do."}
	c.Replacers = []Replacer{{"~", "&thinsp;"}}
	c.ATProto.DID = "did:plc:1234"
	c.ATProto.Password = "your password"
	c.ATProto.PublicationRKey = "foobar"
	c.ATProto.DisplayName = "foobar"
}

var defaultMarkdownOption markdown.Option

func LoadConfig(p string) *Config {
	b, err := os.ReadFile(p)
	var cfg Config
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Error("reading config file", "error", err)
			return nil
		}
		slog.Warn("config file not found", "path", p)
		slog.Info("creating a new config file", "path", p)
		cfg.DefaultValues()
		b, err = toml.Marshal(&cfg)
		if err != nil {
			slog.Error("marshalling config file", "error", err)
			return nil
		}
		err = os.WriteFile(p, b, 0660)
		if err != nil {
			slog.Error("writing config file", "error", err, "path", p)
		} else {
			slog.Info("config file created", "path", p)
		}
		return nil
	}
	err = toml.Unmarshal(b, &cfg)
	if err != nil {
		slog.Error("unmarshalling config file", "error", err)
		return nil
	}
	if len(cfg.AdminPassword) == 0 {
		cfg.AdminPassword = os.Getenv("SW_ADMIN_PASSWORD")
	}
	defaultMarkdownOption.ImageSource = func(path string) string {
		if strings.HasPrefix(path, "https://") {
			return path
		}
		return "/static/" + strings.TrimPrefix(path, "/")
	}
	defaultMarkdownOption.Replaces = make(map[rune]string, len(cfg.Replacers))
	for _, r := range cfg.Replacers {
		if len(r.Symbol) != 1 {
			slog.Error("invalid symbol in config", "symbol", r.Symbol)
			return nil
		}
		defaultMarkdownOption.Replaces[[]rune(r.Symbol)[0]] = r.Replace
	}
	for _, sec := range cfg.Sections {
		err = sec.Init(sec.Folder)
		if err != nil {
			slog.Error("cannot load section", "error", err, "name", sec.Name)
			return nil
		}
	}
	return &cfg
}
