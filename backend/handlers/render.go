package handlers

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"math/rand/v2"
	"net/http"
	"strings"

	"anhgelus.world/small-web/backend"
	"github.com/pelletier/go-toml/v2"
)

//go:embed templates
var templates embed.FS

type PageKind uint8

const (
	article PageKind = iota
	home
	root
)

type Any = any

type PageData = struct {
	Any
	Description string
	URI         string
	Image       backend.ImageHeader
	Tags        []string
	PubDate     toml.LocalDate
	Title       string
	Section     *backend.Section
	Linked      template.HTML
}

type Data struct {
	// global
	backend.Logo
	PageData
	Domain          string
	SiteName        string
	Language        string
	Links           []backend.Link
	PageTitle       string
	PageDescription string
	First           bool
	Kind            PageKind
	PubDate         string
	quotes          []string
}

func (d *Data) Quote() string {
	return d.quotes[rand.IntN(len(d.quotes))]
}

func funcMap(ctx context.Context) template.FuncMap {
	cfg := backend.ContextConfig(ctx)
	return template.FuncMap{
		"static": getStatic,
		"fullStatic": func(path string) string {
			s := getStatic(path)
			if strings.HasPrefix(s, "https://") {
				return s
			}
			return "https://" + cfg.Domain + s
		},
		"asset":       func(path string) backend.AssetData { return getAsset(ctx, path) },
		"kindArticle": func() PageKind { return article },
		"kindRoot":    func() PageKind { return root },
		"kindHome":    func() PageKind { return home },
	}
}

func render(ctx context.Context, r *http.Request, w http.ResponseWriter, file string, pageData PageData) error {
	t, err := template.New("base.html").Funcs(funcMap(ctx)).ParseFS(
		templates,
		"templates/base.html",
		"templates/components.html",
		"templates/"+file+".html",
	)
	if err != nil {
		panic(err)
	}
	cfg := backend.ContextConfig(ctx)
	var data Data
	data.quotes = cfg.Quotes
	data.Language = cfg.Language
	data.Logo = cfg.Logo
	data.Links = cfg.Links
	data.SiteName = cfg.Name
	data.Domain = cfg.Domain
	data.First = !strings.HasPrefix(r.Header.Get("Referer"), "https://"+cfg.Domain)
	pTitle := pageData.Title
	sec := pageData.Section
	if len(pTitle) != 0 {
		if sec != nil {
			pTitle += " - " + sec.TitleName + " entry"
		}
		data.PageTitle = pTitle + " - " + data.SiteName
	} else {
		data.PageTitle = data.SiteName
	}
	pDesc := pageData.Description
	if len(pDesc) == 0 {
		data.PageDescription = cfg.Description
	} else {
		data.PageDescription = pDesc
	}
	uri := pageData.URI
	switch {
	case len(uri) == 0:
		data.Kind = home
	case sec != nil:
		data.Kind = article
	default:
		data.Kind = root
	}
	data.PubDate = pageData.PubDate.String()
	data.PageData = pageData
	return t.Execute(w, &data)
}

type RSSData struct {
	Domain      string
	Language    string
	Title       string
	Description string
	URI         string
	Items       []*backend.Article
}

func renderRSS(ctx context.Context, w http.ResponseWriter, data RSSData) error {
	cfg := backend.ContextConfig(ctx)
	data.Domain = cfg.Domain
	data.Language = cfg.Language
	t, err := template.New("rss.xml").
		Funcs(funcMap(ctx)).
		ParseFS(templates, "templates/rss.xml")
	if err != nil {
		panic(err)
	}
	return t.Execute(w, data)
}

func getStatic(path string) string {
	if strings.HasPrefix(path, "https://") {
		return path
	}
	return "/static/" + strings.TrimPrefix(path, "/")
}

var Assets = map[string]backend.AssetData{}

func getAsset(ctx context.Context, path string) backend.AssetData {
	asset, ok := Assets[path]
	if ok && !backend.ContextDebug(ctx) {
		return asset
	}
	asset = backend.AssetData{}
	logger := backend.ContextLogger(ctx)
	var b []byte
	if strings.HasPrefix(path, "https://") {
		asset.Src = path
		resp, err := http.Get(path)
		if err != nil {
			logger.Warn("get remote asset", "error", err)
			return asset
		}
		defer resp.Body.Close()
		b, err = io.ReadAll(resp.Body)
		if err != nil {
			logger.Warn("read remote asset", "error", err)
			return asset
		}
	} else {
		asset.Src = fmt.Sprintf("/assets/%s", path)
		aFS := backend.ContextAssetsFS(ctx)
		var err error
		b, err = fs.ReadFile(aFS, path)
		if err != nil {
			logger.Warn("read asset", "error", err)
			return asset
		}
	}
	sum := sha256.Sum256(b)
	checksum := base64.StdEncoding.EncodeToString(sum[:])
	asset.Checksum = fmt.Sprintf("sha256-%s", checksum)
	Assets[path] = asset
	return asset
}
