package main

import (
	"context"
	"database/sql"
	"embed"
	"flag"
	"io"
	"log/slog"
	"log/syslog"
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
	atp "anhgelus.world/small-web/atproto"
	"anhgelus.world/small-web/backend"
	"anhgelus.world/small-web/backend/handlers"
	"anhgelus.world/small-web/backend/storage"
	"anhgelus.world/xrpc"
	"anhgelus.world/xrpc/atproto"
	"anhgelus.world/xrpc/server"
	"github.com/nyttikord/logos"
)

//go:embed dist
var embeds embed.FS

var (
	configFile = "config.toml"
	address    = ":8000"
	dev        = false
	sync       = false
	fcgi       = false
	toSyslog   = false
	verbose    = false
)

func init() {
	flag.StringVar(&configFile, "config", configFile, "config file")
	flag.StringVar(&address, "address", address, "address to listen to")
	flag.BoolVar(&dev, "dev", dev, "development mode")
	flag.BoolVar(&sync, "sync", sync, "sync everything with stored data in ATProto PDS")
	flag.BoolVar(&fcgi, "fcgi", fcgi, "use fcgi")
	flag.BoolVar(&toSyslog, "syslog", toSyslog, "log to syslog instead of stderr")
	flag.BoolVar(&verbose, "v", verbose, "increase verbosity")
}

func main() {
	flag.Parse()

	opts := &logos.Options{Level: slog.LevelInfo}
	if dev || verbose {
		opts.Level = slog.LevelDebug
	}

	var h slog.Handler
	var err error
	if toSyslog {
		h, err = logos.NewSyslog("small-web", syslog.LOG_USER, opts)
	} else {
		h = logos.NewColor(os.Stderr, opts)
	}
	if err != nil {
		panic(err)
	}

	slog.SetDefault(slog.New(h))

	cfg := backend.LoadConfig(configFile)
	if cfg == nil {
		slog.Info("exiting")
		os.Exit(1)
	}

	ctx, cancelMigration := context.WithTimeout(
		context.Background(),
		15*time.Second)
	defer cancelMigration()
	db := storage.ConnectDatabase(cfg.Database)
	defer db.Close()
	err = storage.RunMigration(ctx, db)
	if err != nil {
		panic(err)
	}

	ctx, cancelNext := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer cancelNext()

	did, err := atproto.ParseDID(cfg.ATProto.DID)
	if err != nil {
		panic(err)
	}

	if sync {
		syncDocuments(ctx, db, cfg, did)
	}

	assetsFS := handlers.UsableEmbedFS("dist", embeds)
	if dev {
		assetsFS = os.DirFS("dist")
	}

	r := ljus.New()

	r.Use(func(next ljus.Handler, w *ljus.StatusWriter, r *http.Request) {
		ctx := r.Context()
		var ip string
		if fcgi {
			ip = r.RemoteAddr
		} else {
			ip := r.Header.Get("X-Real-Ip")
			if ip == "" {
				ip = r.Header.Get("X-Forwarded-For")
			}
			if ip == "" {
				ip = r.RemoteAddr
			}
		}
		if strings.Contains(ip, ":") {
			ip = strings.Split(ip, ":")[0]
		}
		next(w, r.WithContext(backend.SetContextIP(ctx, ip)))
	})

	r.Use(middleware.Log(slog.Default(), false, false))
	if !dev {
		r.Use(middleware.SecurityHeaders(cfg.Domain, 24*time.Hour))
	}
	r.Use(backend.ContextMiddleware(handlers.Assets, cfg, dev, db),
		backend.RateLimitMiddleware(),
		storage.StatsMiddleware())

	r.Handle(ljus.NewRoute("/", handlers.NotFound()).SetName("not-found"))

	r.Handle(ljus.NewRoute(
		"GET /.well-known/site.standard.publication",
		site.HandlePublicationVerification(did, cfg.ATProto.PublicationRKey)).
		SetName("atproto-verification"))

	r.Handle(ljus.NewRoute("GET /{$}", handlers.Home()).SetName("root"))
	r.Handle(ljus.NewRoute("GET /rss", handlers.RSS()).SetName("rss"))
	r.Handle(ljus.NewRouteFunc("GET /{any}", func(w http.ResponseWriter, req *http.Request) {
		v := req.PathValue("any")
		if strings.HasSuffix(v, ".txt") {
			handlers.TxtFiles().ServeHTTP(w, req)
			return
		}
		handlers.Root().ServeHTTP(w, req)
	}).SetName("any-catcher"))
	r.Handle(ljus.NewRoute("GET /admin", handlers.Admin()).SetName("admin"))

	for _, sec := range cfg.Sections {
		g := ljus.NewGroup("GET /" + sec.Name + "/")
		g.Add(ljus.NewRoute("GET /{$}", handlers.SectionHome(sec)).SetName("root"))
		g.Add(ljus.NewRoute("/{slug}", handlers.SectionArticle(sec)).SetName("article"))
		g.Add(ljus.NewRoute("GET /rss", handlers.SectionRSS(sec)).SetName("rss"))
		r.Handle(g.SetName("section " + sec.Name))
	}

	r.Handle(handlers.StaticFiles("/assets", assetsFS),
		handlers.StaticFiles("/static", os.DirFS(cfg.PublicFolder)))

	slog.Info("starting http server")

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	ctx = backend.SetContextAssetsFS(ctx, assetsFS)

	var l net.Listener
	if strings.HasPrefix(address, "/") {
		l, err = ljus.ListenSocket(address, 0o666)
	} else {
		l, err = net.Listen("tcp", address)
	}
	if err != nil {
		panic(err)
	}
	if fcgi {
		err = r.ServeFastCGI(ctx, l)
	} else {
		err = r.Serve(ctx, l)
	}
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
	art *backend.Article,
) {
	contribs := make([]*site.Contributor, 1, len(art.Contributors)+1)
	contribs[0] = &site.Contributor{
		DID:         did,
		Role:        "Autheur",
		DisplayName: cfg.ATProto.DisplayName,
	}
	for k, v := range art.Contributors {
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
	imgPath := &art.Image.Src
	if v, ok := docs[art.URI]; ok && v.ImageUploaded {
		imgPath = nil
	}
	res, rkey, err := s.PublishDoc(
		ctx,
		client,
		art.Title,
		art.URI,
		art.PubLocalDate.AsTime(time.Local),
		art.Description,
		imgPath,
		art.Tags,
		contribs)
	if err != nil {
		panic(err)
	}
	err = storage.SetPublishedDocument(ctx, db, storage.PublishedDocument{
		Path:          art.URI,
		RecordKey:     rkey,
		CID:           res.CID,
		ImageUploaded: imgPath != nil,
	})
	if err != nil {
		panic(err)
	}
}

func xrpcClient(ctx context.Context, cfg *backend.Config, did *atproto.DID) xrpc.Client {
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
	return client
}

func syncDocuments(ctx context.Context, db *sql.DB, cfg *backend.Config, did *atproto.DID) {
	docs, err := storage.PublishedDocuments(ctx, db)
	if err != nil {
		panic(err)
	}

	files := os.DirFS(cfg.PublicFolder)

	client := xrpcClient(ctx, cfg, did)
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
		for _, art := range sec.Articles() {
			publishDoc(ctx, client, db, docs, cfg, did, s, art)
		}
		slog.Info("syncing done", "section", sec.Name)
	}
	slog.Info("syncing done", "rkey", cfg.ATProto.PublicationRKey)
}
