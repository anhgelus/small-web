package backend

import (
	"bytes"
	"os"

	"github.com/pelletier/go-toml/v2"
)

func Parse(filePath string) (*Article, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var art Article
	data, _, ok := bytes.Cut(b, []byte("---"))
	if ok {
		err = toml.Unmarshal(data, &art)
		if err != nil {
			return nil, err
		}
		art.filePath = filePath
	}
	return &art, nil
}
