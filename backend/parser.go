package backend

import (
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"strings"

	"git.anhgelus.world/anhgelus/small-world/markdown"
	"github.com/pelletier/go-toml/v2"
)

type EntryInfo struct {
	Title        string         `toml:"title"`
	Description  string         `toml:"description"`
	Img          image          `toml:"image"`
	PubLocalDate toml.LocalDate `toml:"publication_date"`
}

func parse(b []byte, info *EntryInfo, d *data) (template.HTML, bool) {
	var dd string
	splits := strings.SplitN(string(b), "---", 2)
	if len(splits) == 2 && info != nil {
		err := toml.Unmarshal([]byte(splits[0]), info)
		if err != nil {
			slog.Warn("parsing entry info", "error", err)
		} else {
			dd = splits[1]
		}
	} else {
		dd = string(b)
	}
	content, err := markdown.Parse(dd, &markdown.Option{ImageSource: getStatic})
	var errMd *markdown.ParseError
	errors.As(err, &errMd)
	if errMd != nil {
		slog.Error("parsing markdown")
		fmt.Println(errMd.Pretty())
		return "", false
	}
	d.PageDescription = info.Description
	d.title = info.Title
	d.Image = info.Img.Src
	return content, true
}
