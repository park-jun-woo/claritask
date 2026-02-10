package webui

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"strings"
)

//go:embed dist/*
var staticFiles embed.FS

// setCacheHeaders sets appropriate cache headers based on path.
// Hashed assets (under /assets/) get long-term cache; everything else gets no-cache.
func setCacheHeaders(w http.ResponseWriter, path string) {
	if strings.HasPrefix(path, "/assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	}
}

// Handler returns an http.Handler that serves the web UI.
// If dir is non-empty, serves from the filesystem (hot-reload without restart).
// Otherwise serves from the embedded dist.
func Handler(dir string) http.Handler {
	var fileSystem http.FileSystem
	if dir != "" {
		fileSystem = http.Dir(dir)
	} else {
		distFS, err := fs.Sub(staticFiles, "dist")
		if err != nil {
			panic("webui: failed to access dist: " + err.Error())
		}
		fileSystem = http.FS(distFS)
	}

	fileServer := http.FileServer(fileSystem)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Try to serve the exact file
		if path != "/" && !strings.HasSuffix(path, "/") {
			if dir != "" {
				// Filesystem: check with os.Stat
				if _, err := os.Stat(dir + path); err == nil {
					setCacheHeaders(w, path)
					fileServer.ServeHTTP(w, r)
					return
				}
			} else {
				// Embedded: check with fs.Stat
				distFS, _ := fs.Sub(staticFiles, "dist")
				if _, err := fs.Stat(distFS, strings.TrimPrefix(path, "/")); err == nil {
					setCacheHeaders(w, path)
					fileServer.ServeHTTP(w, r)
					return
				}
			}
		}

		// SPA fallback: serve index.html for all other routes
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
