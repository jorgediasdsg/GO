package api

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

// DTOs
type ShortenRequest struct {
	URL string `json:"url"`
}
type ShortenResponse struct {
	Error    string `json:"error,omitempty"`
	ShortURL string `json:"short_url,omitempty"`
}

func NewHandler() http.Handler {
	r := chi.NewMux()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	r.Post("api/shorten", handlePost)
	r.Get("/{code}", handleGet)

	return r
}

func handlePost(w http.ResponseWriter, r *http.Request) {

}
func handleGet(w http.ResponseWriter, r *http.Request) {}
