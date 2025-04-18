/**
 * This work is licensed under Apache License, Version 2.0 or later.
 * Please read and understand latest version of Licence.
 */
package webserver

import (
	"compress/gzip"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/kazimsarikaya/assesmentbarkinrl/internal/config"
	"github.com/kazimsarikaya/assesmentbarkinrl/internal/logger"
	"github.com/kazimsarikaya/assesmentbarkinrl/internal/static"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func StartWebServer() (*http.Server, error) {
	var listener net.Listener
	var err error

	listener, err = net.Listen("tcp", fmt.Sprintf(":%d", config.GetConfig().GetServerPort()))

	if err != nil {
		slog.Error("Error starting TCP listener", "error", err)
		return nil, err
	}

	if config.GetConfig().GetLocalStaticPath() != "" || config.GetConfig().GetDebug() {
		// from folder frontend/dist
		slog.Info("Serving static files from local path", "path", config.GetConfig().GetLocalStaticPath())
		staticHandler = http.FileServer(http.Dir(config.GetConfig().GetLocalStaticPath()))
	} else {
		slog.Info("Serving static files from embedded resources")
		staticHandler = http.FileServer(http.FS(static.Static))
	}

	// Create a router
	r := mux.NewRouter()

	// Subrouter for /api paths
	apiRouter := r.PathPrefix("/api").Subrouter()

	apiRouter.HandleFunc("", ApiHandler).Methods(http.MethodGet, http.MethodPost)

	apiRouter.Use(func(next http.Handler) http.Handler {
		return handlers.CompressHandlerLevel(next, gzip.BestCompression)
	})

	// Static files
	r.PathPrefix("/").HandlerFunc(SPAHandler)

	// 404 middleware with logging using combined logger
	r.NotFoundHandler = handlers.CustomLoggingHandler(
		os.Stdout,
		http.HandlerFunc(NotFoundHandler),
		logger.HttpLogFormater)

	// Logging middleware
	r.Use(func(next http.Handler) http.Handler {
		return handlers.CustomLoggingHandler(os.Stdout, next, logger.HttpLogFormater)
	})

	// Recover middleware
	r.Use(func(next http.Handler) http.Handler {
		return handlers.RecoveryHandler(
			handlers.PrintRecoveryStack(true),
			handlers.RecoveryLogger(logger.DefaultLogger),
		)(next)
	})

	// Proxy headers middleware
	r.Use(handlers.ProxyHeaders)

	h2s := &http2.Server{}
	h2cr := h2c.NewHandler(r, h2s)

	srv := &http.Server{
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      h2cr, // Pass our instance of gorilla/mux in.
		ErrorLog:     logger.DefaultErrorLogger,
	}

	go func() {
		if err := srv.Serve(listener); err != nil {
			slog.Error("Error starting server", "error", err)
		}
	}()

	slog.Info("Web server started")

	return srv, nil
}
