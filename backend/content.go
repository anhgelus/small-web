package backend

import (
	"bytes"
	"os"

	"github.com/pelletier/go-toml/v2"
)

func Parse(cfg *Config, filePath string) (*Article, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	rep := make(map[rune]string, len(cfg.Replacers))
	for _, r := range cfg.Replacers {
		rep[[]rune(r.Symbol)[0]] = r.Replace
	}
	var art Article
	art.Replacers = rep
	data, _, ok := bytes.Cut(b, []byte("---"))
	if ok {
		err = toml.Unmarshal(data, &art)
		if err != nil {
			return nil, err
		}
	}
	art.filePath = filePath
	return &art, nil
}
