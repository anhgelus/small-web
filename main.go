package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"git.anhgelus.world/anhgelus/small-world/backend"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Error("loading .env", "error", err)
	}
}
func main() {
	r := backend.NewRouter()

	slog.Info("starting http server")
	server := &http.Server{Addr: ":8000", Handler: r}

	errChan := make(chan error)
	go startServer(server, errChan)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	ok := true
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
