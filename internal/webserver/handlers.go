/**
 * This work is licensed under Apache License, Version 2.0 or later.
 * Please read and understand latest version of Licence.
 */
package webserver

import (
	"net/http"
	"path/filepath"
	"strings"
)

func SPAHandler(w http.ResponseWriter, r *http.Request) {
	// disable other than GET and HEAD methods
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		sendError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// if requested file not ends with .css .jss .ico .map
	// then page is /
	if !strings.HasSuffix(r.URL.Path, ".ico") &&
		!strings.HasPrefix(r.URL.Path, "/static/") {
		r.URL.Path = "/index.html"
	}

	if strings.HasSuffix(r.URL.Path, "/service-worker.js") {
		w.Header().Set("Service-Worker-Allowed", "/")
	}

	w.Header().Set("X-Frame-Options", "SAMEORIGIN")

	if strings.HasPrefix(r.URL.Path, "/static/") {
		w.Header().Set("Cache-Control", "max-age=86400, public")
	} else {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	}

	// get extension of the file
	ext := filepath.Ext(r.URL.Path)

	// set content type by extension
	switch ext {
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".json":
		w.Header().Set("Content-Type", "application/json")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	case ".map":
		w.Header().Set("Content-Type", "application/json")
	}

	// add .gz to the end of the file name
	r.URL.Path = strings.TrimSuffix(r.URL.Path, "/") + ".gz"
	// add header for gzip
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Vary", "Accept-Encoding")

	staticHandler.ServeHTTP(w, r)
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	sendError(w, "Not Found", http.StatusNotFound)
}

// end of file
