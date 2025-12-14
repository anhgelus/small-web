package backend

import (
	"errors"
	"fmt"
	"html"
	"html/template"
	"log/slog"
	"strings"

	"git.anhgelus.world/anhgelus/small-web/dom"
	"git.anhgelus.world/anhgelus/small-web/markdown"
	"github.com/pelletier/go-toml/v2"
)

type EntryInfo struct {
	Title        string         `toml:"title"`
	Description  template.HTML  `toml:"description"`
	Img          image          `toml:"image"`
	PubLocalDate toml.LocalDate `toml:"publication_date"`
}

func renderLinkFunc(url string) func(string, string) template.HTML {
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
	}
}

func renderLink(content, href, url string) template.HTML {
	return renderLinkFunc(url)(content, href)
}

func parse(b []byte, info *EntryInfo, d *data) (template.HTML, bool) {
	opt := defaultMarkdownOption
	opt.RenderLink = renderLinkFunc(d.URL)

	var dd string
	var err error
	splits := strings.SplitN(string(b), "---", 2)
	if len(splits) == 2 && info != nil {
		err = toml.Unmarshal([]byte(splits[0]), info)
		if err != nil {
			slog.Warn("parsing entry info", "error", err)
		} else {
			info.Description, err = markdown.Parse(string(info.Description), &opt)
			dd = splits[1]
		}
	} else {
		dd = string(b)
	}

	var errMd *markdown.ParseError
	errors.As(err, &errMd)
	var content template.HTML
	if errMd == nil {
		content, err = markdown.Parse(dd, &opt)
		errors.As(err, &errMd)
	}
	if errMd != nil {
		slog.Error("parsing markdown")
		fmt.Println(errMd.Pretty())
		return "", false
	}
	d.PageDescription = html.UnescapeString(string(info.Description))
	d.title = info.Title
	d.Image = info.Img.Src
	return content, true
}
