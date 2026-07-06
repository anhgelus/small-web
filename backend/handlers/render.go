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
)

//go:embed templates
var templates embed.FS

type Data struct {
	// global
	backend.Logo
	Domain   string
	SiteName string
	Language string
	Linked   template.HTML
	Links    []backend.Link
	// page
	PageDescription string
	URL             string
	Image           string
	PubDate         string
	PageTitle       string
	quotes          []string
	Custom          any
}

func (d *Data) Title() string {
	title := d.SiteName
	if len(d.PageTitle) != 0 {
		title = d.PageTitle + " - " + title
	}
	return title
}

func (d *Data) Quote() string {
	return d.quotes[rand.IntN(len(d.quotes))]
}

func render(ctx context.Context, w http.ResponseWriter, file string, data Data) error {
	t, err := template.ParseFS(
		templates,
		"templates/base.html",
		"templates/components.html",
		"templates/"+file+".html",
	)
	if err != nil {
		panic(err)
	}
	cfg := backend.ContextConfig(ctx)
	data.quotes = cfg.Quotes
	data.Language = cfg.Language
	data.Logo = cfg.Logo
	data.Links = cfg.Links
	data.SiteName = cfg.Name
	data.Domain = cfg.Domain
	if len(data.PageDescription) == 0 {
		data.PageDescription = cfg.Description
	}
	t = t.Funcs(template.FuncMap{
		"static": getStatic,
		"fullStatic": func(path string) string {
			s := getStatic(path)
			if strings.HasPrefix(s, "https://") {
				return s
			}
			return "https://" + cfg.Domain + s
		},
		"asset": func(path string) backend.AssetData { return getAsset(ctx, path) },
		"first": func(sl []any) any {
			if len(sl) == 0 {
				return nil
			}
			return sl[0]
		},
		"queue": func(sl []any) any {
			if len(sl) < 2 {
				return nil
			}
			return sl[1:]
		},
		"next":   func(i int) int { return i + 1 },
		"before": func(i int) int { return i - 1 },
	})
	return t.Execute(w, data)
}

func getStatic(path string) string {
	if strings.HasPrefix(path, "https://") {
		return path
	}
	return "/static/" + strings.TrimPrefix(path, "/")
}

var assets = map[string]backend.AssetData{}

func getAsset(ctx context.Context, path string) backend.AssetData {
	asset, ok := assets[path]
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
	assets[path] = asset
	return asset
}
