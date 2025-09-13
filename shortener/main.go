// URL SHORTENER
// A simple URL shortener service in Go
// Features:
// - Shorten URLs
// - Redirect to original URL
// - In-memory storage
// - Basic error handling

package main

import (
	"log/slog"
	"net/http"
	"os"
	"shortener/api"
	"time"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Error running the service", "error", err)
		os.Exit(1)
	}
	slog.Info("All system offline")

}

func run() error {
	handler := api.NewHandler()

	s := http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Addr:         ":8080",
		Handler:      handler,
	}

	slog.Info("Starting server on :8080")
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
