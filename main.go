package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"anhgelus.world/ljus"
	"anhgelus.world/ljus/middleware"
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
)

func init() {
	flag.StringVar(&configFile, "config", configFile, "config file")
	flag.IntVar(&port, "port", port, "server port")
	flag.BoolVar(&dev, "dev", false, "development mode")
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

	for _, sec := range cfg.Sections {
		if ok = sec.Load(cfg); !ok {
			slog.Info("exiting")
			os.Exit(2)
		}
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
