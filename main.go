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

	"git.anhgelus.world/anhgelus/small-world/backend"
	"github.com/joho/godotenv"
)

//go:embed dist
var embeds embed.FS

var (
	configFile = "config.toml"
	port       = 8000
	publicDir  = "public"
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

	if v := os.Getenv("PUBLIC_DIR"); v != "" {
		publicDir = v
	}
	flag.StringVar(&publicDir, "public", publicDir, "public directory")
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

	if ok = backend.LoadLogs(cfg); !ok {
		slog.Info("exiting")
		os.Exit(2)
	}

	r := backend.NewRouter(dev, cfg)

	backend.HandleHome(r)
	backend.HandleRoot(r, cfg)
	backend.HandleLogs(r)

	if dev {
		backend.HandleStaticFiles(r, "/assets", os.DirFS("dist"))
	} else {
		backend.HandleStaticFiles(r, "/assets", backend.UsableEmbedFS("dist", embeds))
	}
	backend.HandleStaticFiles(r, "/static", os.DirFS(publicDir))

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
