package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"git.anhgelus.world/anhgelus/small-web/backend"
	"github.com/joho/godotenv"
)

//go:embed dist
var embeds embed.FS

var (
	configFile = "config.toml"
	port       = 8000
	dev        = false
)

func init() {
	err := godotenv.Load(".env")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Error("loading .env", "error", err)
	}

	if v := os.Getenv("CONFIG_FILE"); v != "" {
		configFile = v
	}
	flag.StringVar(&configFile, "config", configFile, "config file")

	if v := os.Getenv("PORT"); v != "" {
		port, err = strconv.Atoi(v)
		if err != nil {
			panic(err)
		}
	}
	flag.IntVar(&port, "port", port, "server port")
	flag.BoolVar(&dev, "dev", false, "development mode")
}

func main() {
	flag.Parse()

	backend.SetupLogger(dev)

	cfg, ok := backend.LoadConfig(configFile)
	if !ok {
		slog.Info("exiting")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db := backend.ConnectDatabase(cfg)
	defer db.Close()
	err := backend.RunMigration(ctx, db)
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

	r := backend.NewRouter(dev, cfg, db, assetsFS)

	backend.HandleHome(r)
	backend.HandleRoot(r, cfg)
	for _, sec := range cfg.Sections {
		sec.Handle(r)
	}
	backend.Handle404(r)

	backend.HandleAdmin(r)

	backend.HandleStaticFiles(r, "/assets", assetsFS)
	backend.HandleStaticFiles(r, "/static", os.DirFS(cfg.PublicFolder))

	slog.Info("starting http server")
	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: r}

	errChan := make(chan error)
	go startServer(server, errChan)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	ok = true
	for ok {
		select {
		case err := <-errChan:
			slog.Error("http server running", "error", err)
			slog.Info("restarting the server")
			go startServer(server, errChan)
		case <-sc:
			err := server.Shutdown(context.Background())
			if err != nil {
				slog.Error("closing http server", "error", err)
			}
			ok = false
		}
	}
	slog.Info("http server stopped")
}

func startServer(server *http.Server, errChan chan error) {
	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		errChan <- err
	}
}
