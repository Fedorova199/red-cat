package handlers

import (
	"context"
	"net/http"

	"github.com/Fedorova199/red-cat/internal/app/storage"
)

type Storage interface {
	Get(ctx context.Context, id int) (storage.CreateURL, error)
	Set(ctx context.Context, model storage.CreateURL) (int, error)
	GetByUser(ctx context.Context, userID string) ([]storage.CreateURL, error)
	APIShortenBatch(ctx context.Context, records []storage.ShortenBatch) ([]storage.ShortenBatch, error)
	GetByOriginURL(ctx context.Context, originURL string) (storage.CreateURL, error)
	Ping(ctx context.Context) error
}

type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginURL     string `json:"original_url"`
}

type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type ShortURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
type Middleware interface {
	Handle(next http.HandlerFunc) http.HandlerFunc
}

func Middlewares(handler http.HandlerFunc, middlewares []Middleware) http.HandlerFunc {
	for _, middleware := range middlewares {
		handler = middleware.Handle(handler)
	}

	return handler
}
