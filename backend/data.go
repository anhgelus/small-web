package backend

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	txt "text/template"
)

var (
	regexIsHttp = regexp.MustCompile(`^https?://`)
)

type dataUsable interface {
	SetData(*data)
}

type data struct {
	title           string
	Article         bool
	Domain          string
	URL             string
	Image           string
	PageDescription string
	Name            string
	Links           []Link
	Logo            *Logo
	Quote           string
}

func (d *data) SetData(data *data) {
	*d = *data
}

func (d *data) merge(cfg *Config, r *http.Request) {
	if d.Domain == "" {
		d.Domain = cfg.Domain
	}
	if d.Name == "" {
		d.Name = cfg.Name
	}
	if d.PageDescription == "" {
		d.PageDescription = cfg.Description
	}
	if d.Links == nil {
		d.Links = cfg.Links
	}
	if d.Logo == nil {
		d.Logo = &cfg.Logo
	}
	if d.Quote == "" {
		if cfg.Quotes == nil {
			d.Quote = "Une citation"
		} else {
			d.Quote = cfg.Quotes[rand.Intn(len(cfg.Quotes))]
		}
	}
	if d.Image == "" {
		d.Image = cfg.DefaultImage
	}
	if d.URL == "" {
		if !strings.HasPrefix(r.URL.Path, "/") {
			r.URL.Path = "/" + r.URL.Path
		}
		d.URL = r.URL.Path
	}
}

func (d *data) handleGeneric(w http.ResponseWriter, r *http.Request, name string, custom dataUsable) {
	cfg := r.Context().Value(configKey).(*Config)
	d.merge(cfg, r)
	t, err := template.New("").Funcs(template.FuncMap{
		"static": getStatic,
		"fullStatic": func(path string) string {
			if regexIsHttp.MatchString(path) {
				return path
			}
			return fmt.Sprintf("https://%s/static/%s", cfg.Domain, path)
		},
		"asset": func(path string) *assetData {
			return getAsset(r.Context(), path)
		},
		"next":   func(i int) int { return i + 1 },
		"before": func(i int) int { return i - 1 },
		"first":  templateFirst,
	}).ParseFS(templates, "templates/components.html", fmt.Sprintf("templates/%s.html", name), "templates/base.html")
	if err != nil {
		panic(err)
	}
	exec := "base.html"
	if r.Context().Value(isUpdateKey).(bool) {
		exec = "body"
		w.Header().Set("Updated-Title", url.QueryEscape(d.Title()))
		w.Header().Set("Updated-Quote", url.QueryEscape(d.Quote))
	}
	if custom == nil {
		err = t.ExecuteTemplate(w, exec, d)
	} else {
		custom.SetData(d)
		err = t.ExecuteTemplate(w, exec, custom)
	}
	if err != nil {
		panic(err)
	}
}

func (d *data) handleRSS(w http.ResponseWriter, r *http.Request, custom dataUsable) {
	cfg := r.Context().Value(configKey).(*Config)
	d.merge(cfg, r)
	t, err := txt.New("").Funcs(txt.FuncMap{"first": templateFirst}).ParseFS(templates, "templates/rss.xml")
	if err != nil {
		panic(err)
	}
	r.Header.Set("Content-Type", "application/rss+xml")
	if custom == nil {
		err = t.ExecuteTemplate(w, "rss.xml", d)
	} else {
		custom.SetData(d)
		err = t.ExecuteTemplate(w, "rss.xml", custom)
	}
	if err != nil {
		panic(err)
	}
}

func (d *data) Title() string {
	title := d.Name
	if d.Article {
		title += " - log entry"
	}
	if len(d.title) != 0 {
		title += " - " + d.title
	}
	return title
}

func (d *data) PubDate() string {
	return ""
}

func getStatic(path string) string {
	if regexIsHttp.MatchString(path) {
		return path
	}
	return fmt.Sprintf("/static/%s", path)
}

type assetData struct {
	Src      string
	Checksum string
}

var assets = map[string]*assetData{}

func getAsset(ctx context.Context, path string) *assetData {
	asset, ok := assets[path]
	if ok && !ctx.Value(debugKey).(bool) {
		return asset
	}
	asset = &assetData{}
	var b []byte
	if regexIsHttp.MatchString(path) {
		asset.Src = path
		resp, err := http.Get(path)
		if err != nil {
			slog.Warn("get remote asset", "error", err)
			return asset
		}
		defer resp.Body.Close()
		b, err = io.ReadAll(resp.Body)
		if err != nil {
			slog.Warn("read remote asset", "error", err)
			return asset
		}
	} else {
		asset.Src = fmt.Sprintf("/assets/%s", path)
		aFS := ctx.Value(assetsFSKey).(fs.FS)
		var err error
		b, err = fs.ReadFile(aFS, path)
		if err != nil {
			slog.Warn("read asset", "error", err)
			return asset
		}
	}
	sum := sha256.Sum256(b)
	checksum := base64.StdEncoding.EncodeToString(sum[:])
	asset.Checksum = fmt.Sprintf("sha256-%s", checksum)
	assets[path] = asset
	return asset
}

func templateFirst(a []*Section) *Section {
	if len(a) == 0 {
		return nil
	}
	return a[0]
}
