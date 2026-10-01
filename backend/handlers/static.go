package handlers

import (
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"

	"anhgelus.world/ljus"
	"anhgelus.world/small-web/backend"
)

func StaticFiles(path string, root fs.FS) ljus.Route {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	return ljus.NewRouteFunc(path+"{file...}", func(w http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.RequestURI, "/") {
			NotFound().ServeHTTP(w, req)
			return
		}
		http.StripPrefix(path, http.FileServerFS(root)).ServeHTTP(w, req)
	}).SetName("static-files " + path)
}

func TxtFiles() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		cfg := backend.ContextConfig(ctx)
		logger := backend.ContextLogger(ctx)
		logger.Info("requesting txt file", "User-Agent", r.Header.Get("User-Agent"))
		b, err := os.ReadFile(path.Join(cfg.PublicFolder, r.PathValue("any")))
		if os.IsNotExist(err) {
			NotFound().ServeHTTP(w, r)
			return
		} else if err != nil {
			panic(err)
		}
		_, err = w.Write(b)
		if err != nil {
			panic(err)
		}
	})
}
