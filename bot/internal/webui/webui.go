package webui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed dist/*
var staticFiles embed.FS

// Handler returns an http.Handler that serves the embedded web UI.
// - /api/* requests are not handled (pass through)
// - All other requests serve static files with SPA fallback to index.html
func Handler() http.Handler {
	// Get the dist subdirectory
	distFS, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		panic("webui: failed to access dist: " + err.Error())
	}

	fileServer := http.FileServer(http.FS(distFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Try to serve the exact file
		if path != "/" && !strings.HasSuffix(path, "/") {
			// Check if file exists in embedded FS
			if _, err := fs.Stat(distFS, strings.TrimPrefix(path, "/")); err == nil {
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// SPA fallback: serve index.html for all other routes
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
