package backend

import (
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"strings"

	"git.anhgelus.world/anhgelus/small-web/markdown"
	"github.com/pelletier/go-toml/v2"
)

type EntryInfo struct {
	Title        string         `toml:"title"`
	Description  string         `toml:"description"`
	Img          image          `toml:"image"`
	PubLocalDate toml.LocalDate `toml:"publication_date"`
}

func renderLinkFunc(url string) func(string, string) template.HTML {
	return func(content, href string) template.HTML {
		b := "<a"
		if href == url || (href != "/" && url != "/" && strings.HasPrefix(url, href)) {
			b += ` class="target"`
		}
		if markdown.ExternalLink.MatchString(href) {
			b += ` target="_blank"`
		}
		return template.HTML(fmt.Sprintf(`%s href="%s">%s</a>`, b, href, content))
	}
}

func renderLink(content, href, url string) template.HTML {
	return renderLinkFunc(url)(content, href)
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
	opt := new(markdown.Option)
	opt.ImageSource = getStatic
	opt.RenderLink = renderLinkFunc(d.URL)
	content, err := markdown.Parse(dd, opt)
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
