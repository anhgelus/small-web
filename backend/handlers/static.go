package handlers

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"

	"anhgelus.world/ljus"
	"anhgelus.world/small-web/backend"
)

// httpEmbedFS is an implementation of fs.FS, fs.ReadDirFS and fs.ReadFileFS helping to manage embed.FS for http server
type httpEmbedFS struct {
	embed.FS
	prefix string
}

func (h *httpEmbedFS) Open(name string) (fs.File, error) {
	return h.FS.Open(h.prefix + "/" + name)
}

func (h *httpEmbedFS) ReadFile(name string) ([]byte, error) {
	return h.FS.ReadFile(h.prefix + "/" + name)
}

func (h *httpEmbedFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return h.FS.ReadDir(h.prefix + "/" + name)
}

// UsableEmbedFS converts embed.FS into usable fs.FS
func UsableEmbedFS(folder string, em embed.FS) fs.FS {
	return &httpEmbedFS{
		prefix: strings.Trim(folder, "/"),
		FS:     em,
	}
}

func StaticFiles(path string, root fs.FS) ljus.Route {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	return ljus.NewRouteFunc(path+"{file...}", func(w http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.RequestURI, "/") {
			NotFound(w, req)
			return
		}
		http.StripPrefix(path, http.FileServerFS(root)).ServeHTTP(w, req)
	}).SetName("static-files " + path)
}

func TxtFiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cfg := backend.ContextConfig(ctx)
	logger := backend.ContextLogger(ctx)
	logger.Info("requesting txt file", "User-Agent", r.Header.Get("User-Agent"))
	b, err := os.ReadFile(path.Join(cfg.PublicFolder, r.PathValue("any")))
	if os.IsNotExist(err) {
		NotFound(w, r)
		return
	} else if err != nil {
		panic(err)
	}
	_, err = w.Write(b)
	if err != nil {
		panic(err)
	}
}
