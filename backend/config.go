package backend

import (
	"errors"
	"log/slog"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Domain string `toml:"domain"`
}

func (c *Config) DefaultValues() {
	c.Domain = "example.org"
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
