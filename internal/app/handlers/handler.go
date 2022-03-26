package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/Fedorova199/red-cat/internal/app/storages"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

type Storage interface {
	Get(ctx context.Context, id uint64) (storages.Record, error)
	GetByOriginURL(ctx context.Context, originURL string) (storages.Record, error)
	GetByUser(ctx context.Context, userID string) ([]storages.Record, error)
	Put(ctx context.Context, record storages.Record) (uint64, error)
	PutRecords(ctx context.Context, records []storages.BatchRecord) ([]storages.BatchRecord, error)
	Ping(ctx context.Context) error
}

type Middleware interface {
	Handle(next http.HandlerFunc) http.HandlerFunc
}

type Handler struct {
	*chi.Mux
	Storage Storage
	BaseURL string
}

func (h *Handler) ShowNotFoundPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func NewHandler(storage Storage, baseURL string, middlewares []Middleware) *Handler {
	h := &Handler{
		Mux:     chi.NewMux(),
		Storage: storage,
		BaseURL: baseURL,
	}

	h.Get("/ping", applyMiddlewares(h.Ping(), middlewares))
	h.Get("/{id}", applyMiddlewares(h.GetOriginalURL(), middlewares))
	h.Get("/api/user/urls", applyMiddlewares(h.GetAllUrls(), middlewares))
	h.Post("/", applyMiddlewares(h.ShortenURL(), middlewares))
	h.Post("/api/shorten", applyMiddlewares(h.APIShortenURL(), middlewares))
	h.Post("/api/shorten/batch", applyMiddlewares(h.APIShortenBatch(), middlewares))
	h.NotFound(applyMiddlewares(h.ShowNotFoundPage(), middlewares))

	return h
}

func applyMiddlewares(handler http.HandlerFunc, middlewares []Middleware) http.HandlerFunc {
	for _, middleware := range middlewares {
		handler = middleware.Handle(handler)
	}

	return handler
}

func (h *Handler) ShortenURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userCookie, err := r.Cookie("user_id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		url := string(b)
		id, err := h.Storage.Put(r.Context(), storages.Record{
			User: userCookie.Value,
			URL:  url,
		})

		if err != nil {
			var pge *pgconn.PgError
			if errors.As(err, &pge) && pge.Code == pgerrcode.UniqueViolation {
				record, err := h.Storage.GetByOriginURL(r.Context(), url)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				resultURL := h.BaseURL + "/" + strconv.FormatUint(record.ID, 10)
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(resultURL))
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resultURL := h.BaseURL + "/" + strconv.FormatUint(id, 10)

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(resultURL))
	}
}

func (h *Handler) GetOriginalURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawID := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(rawID, 10, 64)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		record, err := h.Storage.Get(r.Context(), id)

		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Location", record.URL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

func (h *Handler) APIShortenURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		request := Request{}
		if err := json.Unmarshal(b, &request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		userCookie, err := r.Cookie("user_id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		id, err := h.Storage.Put(r.Context(), storages.Record{
			User: userCookie.Value,
			URL:  request.URL,
		})

		if err != nil {
			var pge *pgconn.PgError
			if errors.As(err, &pge) && pge.Code == pgerrcode.UniqueViolation {
				record, err := h.Storage.GetByOriginURL(r.Context(), request.URL)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				res, err := h.formatResult(record.ID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				w.Write(res)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res, err := h.formatResult(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(res)
	}
}

func (h *Handler) formatResult(id uint64) ([]byte, error) {
	resultURL := h.BaseURL + "/" + strconv.FormatUint(id, 10)
	response := Response{Result: resultURL}
	return json.Marshal(response)
}

type ShortenURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (h *Handler) GetAllUrls() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCookie, err := r.Cookie("user_id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		records, err := h.Storage.GetByUser(r.Context(), userCookie.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(records) == 0 {
			http.Error(w, "Not found", http.StatusNoContent)
			return
		}

		var shortenUrls []ShortenURL
		for _, record := range records {
			shortenUrls = append(shortenUrls, ShortenURL{
				ShortURL:    h.BaseURL + "/" + strconv.FormatUint(record.ID, 10),
				OriginalURL: record.URL,
			})
		}

		res, err := json.Marshal(shortenUrls)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	}
}

func (h *Handler) Ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h.Storage.Ping(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginURL     string `json:"original_url"`
}

type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func (h *Handler) APIShortenBatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var batchRequests []BatchRequest
		if err := json.Unmarshal(b, &batchRequests); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		userCookie, err := r.Cookie("user_id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var batchRecords []storages.BatchRecord
		for _, batchRequest := range batchRequests {
			batchRecords = append(batchRecords, storages.BatchRecord{
				User:          userCookie.Value,
				URL:           batchRequest.OriginURL,
				CorrelationID: batchRequest.CorrelationID,
			})
		}

		batchRecords, err = h.Storage.PutRecords(r.Context(), batchRecords)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var batchResponses []BatchResponse
		for _, batchRecord := range batchRecords {
			batchResponses = append(batchResponses, BatchResponse{
				CorrelationID: batchRecord.CorrelationID,
				ShortURL:      h.BaseURL + "/" + strconv.FormatUint(batchRecord.ID, 10),
			})
		}

		res, err := json.Marshal(batchResponses)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(res)
	}
}
