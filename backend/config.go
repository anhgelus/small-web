package backend

import (
	"errors"
	"log/slog"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Link struct {
	Name string `toml:"name"`
	URL  string `toml:"url"`
}

type Logo struct {
	Header  string `toml:"header"`
	Favicon string `toml:"favicon"`
}

type Config struct {
	Domain      string `toml:"domain"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Links       []Link `toml:"links"`
	Logo        Logo   `toml:"logo"`
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
}

func LoadConfig(path string) (*Config, bool) {
	b, err := os.ReadFile(path)
	var config Config
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
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
	return &config, true
}
