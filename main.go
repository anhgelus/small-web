package main

import (
	"context"
	"database/sql"
	"embed"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	site "anhgelus.world/goat-site"
	"anhgelus.world/ljus"
	"anhgelus.world/ljus/middleware"
	"anhgelus.world/xrpc"
	"anhgelus.world/xrpc/atproto"
	"anhgelus.world/xrpc/server"
	atp "git.anhgelus.world/anhgelus/small-web/atproto"
	"git.anhgelus.world/anhgelus/small-web/backend"
	"git.anhgelus.world/anhgelus/small-web/backend/common"
	"git.anhgelus.world/anhgelus/small-web/backend/storage"
)

//go:embed dist
var embeds embed.FS

var (
	configFile = "config.toml"
	port       = 8000
	dev        = false
	sync       = false
	publish    = ""
)

func init() {
	flag.StringVar(&configFile, "config", configFile, "config file")
	flag.IntVar(&port, "port", port, "server port")
	flag.BoolVar(&dev, "dev", dev, "development mode")
	flag.BoolVar(&sync, "sync", sync, "sync everything with stored data in ATProto PDS")
	flag.StringVar(&publish, "publish", publish, "publish an article on ATProto")
}

func main() {
	flag.Parse()

	cfg, ok := backend.LoadConfig(configFile)
	if !ok {
		slog.Info("exiting")
		os.Exit(1)
	}

	ctx, cancelMigration := context.WithTimeout(
		context.Background(),
		15*time.Second)
	defer cancelMigration()
	db := storage.ConnectDatabase(cfg.Database)
	defer db.Close()
	err := storage.RunMigration(ctx, db)
	if err != nil {
		panic(err)
	}

	docs, err := storage.PublishedDocuments(ctx, db)
	if err != nil {
		panic(err)
	}

	for _, sec := range cfg.Sections {
		if ok = sec.Load(docs); !ok {
			slog.Info("exiting")
			os.Exit(2)
		}
	}

	did, err := atproto.ParseDID(cfg.ATProto.DID)
	if err != nil {
		panic(err)
	}

	ctx, cancelNext := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer cancelNext()

	var client xrpc.Client = xrpc.NewClient(
		http.DefaultClient,
		atproto.NewDirectory(http.DefaultClient, net.DefaultResolver),
		"Small Web 1.0")
	doc, err := client.Directory().ResolveDID(ctx, did)
	if err != nil {
		panic(err)
	}
	pds, _ := doc.PDS()
	res, err := server.CreateSession(
		ctx,
		client,
		pds,
		server.CreateSessionRequest{Identifier: cfg.ATProto.DID, Password: cfg.ATProto.Password})
	if err != nil {
		panic(err)
	}
	client = res.Client
	err = server.RefreshSession(ctx, client.(*xrpc.AuthClient))
	if err != nil {
		panic(err)
	}

	files := os.DirFS(cfg.PublicFolder)

	if sync {
		var logo []byte
		var name string
		if strings.HasPrefix(cfg.Logo.Favicon, "https://") {
			raw := strings.Split(cfg.Logo.Favicon, "/")
			name = raw[len(raw)-1]
			resp, err := http.Get(cfg.Logo.Favicon)
			if err != nil {
				panic(err)
			}
			logo, err = io.ReadAll(resp.Body)
		} else {
			f, err := files.Open(cfg.Logo.Favicon)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			logo, err = io.ReadAll(f)
			if err != nil {
				panic(err)
			}
			name = cfg.Logo.Favicon
		}
		blob, err := xrpc.UploadBlob(
			ctx, client, mime.TypeByExtension("."+strings.Split(name, ".")[1]), logo)
		if err != nil {
			panic(err)
		}
		u, _ := url.Parse("https://" + cfg.Domain)
		s, err := atp.CreateSite(ctx,
			client,
			files,
			did,
			cfg.ATProto.PublicationRKey,
			&site.Publication{
				URL:         u,
				Name:        cfg.Name,
				Icon:        blob,
				Description: &cfg.Description,
				Preferences: &site.Preferences{ShowInDiscover: true},
			})
		if err != nil {
			panic(err)
		}
		for _, sec := range cfg.Sections {
			for _, data := range sec.Data {
				publishDoc(ctx, client, db, docs, cfg, did, s, data.URL, &data.EntryInfo)
			}
			slog.Info("syncing done", "section", sec.Name)
		}
		slog.Info("syncing done", "rkey", cfg.ATProto.PublicationRKey)
		return
	}
	if publish != "" {
		s, err := atp.LoadSite(ctx, client, files, did, cfg.ATProto.PublicationRKey)
		if err != nil {
			panic(err)
		}
		info, err := backend.Publish(publish)
		if err != nil {
			panic(err)
		}
		publishDoc(ctx, client, db, docs, cfg, did, s, publish, info)
		slog.Info("publishing done", "path", publish)
		return
	}

	assetsFS := backend.UsableEmbedFS("dist", embeds)
	if dev {
		assetsFS = os.DirFS("dist")
	}

	r := ljus.New()

	r.Use(middleware.Log(slog.Default(), false, false))
	if !dev {
		r.Use(middleware.SecurityHeaders(cfg.Domain, 24*time.Hour))
	}
	r.Use(backend.ContextMiddleware(cfg, dev, db),
		backend.RateLimitMiddleware(),
		backend.DumbBotMiddleware(),
		backend.StatsMiddleware())

	r.NotFoundHandler = http.HandlerFunc(backend.NotFoundHandler)

	r.Handle(ljus.NewRouteFunc(
		"GET /.well-known/site.standard.publication",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(
				"at://" + cfg.ATProto.DID + "/site.standard.publication/" + cfg.ATProto.PublicationRKey.String()))
		}).SetName("atproto-verification"))

	r.Handle(ljus.NewRouteFunc("GET /{$}", backend.HomeHandler).SetName("root"))
	r.Handle(ljus.NewRouteFunc(
		"GET /rss",
		backend.GenericRSSHandler,
	).SetName("rss"))
	r.Handle(ljus.NewRouteFunc("GET /{any}", func(w http.ResponseWriter, req *http.Request) {
		v := req.PathValue("any")
		if strings.HasSuffix(v, ".txt") {
			backend.TxtFilesHandler(w, req)
			return
		}
		backend.GenericRootHandler(w, req)
	}).SetName("any-catcher"))
	r.Handle(ljus.NewRouteFunc("GET /admin", backend.AdminHandler).SetName("admin"))

	for _, sec := range cfg.Sections {
		g := ljus.NewGroup("GET /" + sec.Name + "/")
		g.Add(ljus.NewRouteFunc("/", sec.RootHandler).SetName("root"))
		g.Add(ljus.NewRouteFunc("/{slug}", sec.Handler).SetName("article"))
		g.Add(ljus.NewRouteFunc("GET /rss", sec.RSSHandler).SetName("rss"))
		g.Add(ljus.NewRouteFunc("GET /rss/", sec.RSSHandler).SetName("rss"))
		r.Handle(g.SetName("section " + sec.Name))
	}

	r.Handle(backend.StaticFilesHandler("/assets", assetsFS),
		backend.StaticFilesHandler("/static", os.DirFS(cfg.PublicFolder)))

	slog.Info("starting http server")

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	ctx = common.SetContextAssetsFS(ctx, assetsFS)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	err = r.Serve(ctx, l)
	select {
	case <-ctx.Done():
	default:
		panic(err)
	}
	slog.Info("http server stopped")
}

func publishDoc(
	ctx context.Context,
	client xrpc.Client,
	db *sql.DB,
	docs map[string]storage.PublishedDocument,
	cfg *backend.Config,
	did *atproto.DID,
	s *atp.Site,
	path string,
	info *backend.EntryInfo,
) {
	contribs := make([]*site.Contributor, 1, len(info.Contributors)+1)
	contribs[0] = &site.Contributor{
		DID:         did,
		Role:        "Autheur",
		DisplayName: cfg.ATProto.DisplayName,
	}
	for k, v := range info.Contributors {
		d, err := atproto.ParseDID(v.DID)
		if err != nil {
			panic(d)
		}
		contribs = append(contribs, &site.Contributor{
			DisplayName: k,
			Role:        v.Role,
			DID:         d,
		})
	}
	imgPath := &info.Img.Src
	if v, ok := docs[path]; ok && v.ImageUploaded {
		imgPath = nil
	}
	res, rkey, err := s.PublishDoc(
		ctx,
		client,
		info.Title,
		path,
		info.PubLocalDate.AsTime(time.Local),
		info.Description,
		imgPath,
		info.Tags,
		contribs)
	if err != nil {
		panic(err)
	}
	err = storage.SetPublishedDocument(ctx, db, storage.PublishedDocument{
		Path:          path,
		RecordKey:     rkey,
		CID:           res.CID,
		ImageUploaded: imgPath != nil,
	})
	if err != nil {
		panic(err)
	}
}
