package interfaces

import (
	"context"
	"net/http"

	"github.com/Fedorova199/red-cat/internal/app/storage"
)

type Storage interface {
	Get(ctx context.Context, id int) (storage.CreateURL, error)
	GetOriginURL(ctx context.Context, originURL string) (storage.CreateURL, error)
	GetUser(ctx context.Context, userID string) ([]storage.CreateURL, error)
	Set(ctx context.Context, createURL storage.CreateURL) (int, error)
	PutBatch(ctx context.Context, shortBatch []storage.ShortenBatch) ([]storage.ShortenBatch, error)
	Ping(ctx context.Context) error
	DeleteURLs(ctx context.Context, ids []int) error
}

type Middleware interface {
	Handle(next http.HandlerFunc) http.HandlerFunc
}
