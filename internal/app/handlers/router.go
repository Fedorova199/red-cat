package handlers

import (
	"net/http"

	"github.com/Fedorova199/red-cat/internal/interfaces"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	*chi.Mux
	Storage interfaces.Storage
	BaseURL string
}

func NewHandler(storage interfaces.Storage, baseURL string, middlewares []interfaces.Middleware) *Handler {
	router := &Handler{
		Mux:     chi.NewMux(),
		Storage: storage,
		BaseURL: baseURL,
	}

	router.Get("/ping", Middlewares(router.PingHandler, middlewares))
	router.Get("/{id}", Middlewares(router.GetHandler, middlewares))
	router.Get("/api/user/urls", Middlewares(router.GetUrlsHandler, middlewares))
	router.Post("/", Middlewares(router.PostHandler, middlewares))
	router.Post("/api/shorten", Middlewares(router.JSONHandler, middlewares))
	router.Post("/api/shorten/batch", Middlewares(router.PostAPIShortenBatchHandler, middlewares))
	router.Delete("/api/user/urls", Middlewares(router.DeleteUrls, middlewares))

	return router
}

func Middlewares(handler http.HandlerFunc, middlewares []interfaces.Middleware) http.HandlerFunc {
	for _, middleware := range middlewares {
		handler = middleware.Handle(handler)
	}

	return handler
}
