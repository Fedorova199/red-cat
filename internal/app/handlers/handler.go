package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Fedorova199/red-cat/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

type Handler struct {
	*chi.Mux
	Storage Storage
	BaseURL string
}

func NewHandler(storage Storage, baseURL string, middlewares []Middleware) *Handler {
	router := &Handler{
		Mux:     chi.NewMux(),
		Storage: storage,
		BaseURL: baseURL,
	}
	router.Get("/{id}", Middlewares(router.GETHandler, middlewares))
	router.Get("/api/user/urls", Middlewares(router.GetUrlsHandler, middlewares))
	router.Get("/ping", Middlewares(router.PingHandler, middlewares))
	router.Post("/", Middlewares(router.POSTHandler, middlewares))
	router.Post("/api/shorten", Middlewares(router.JSONHandler, middlewares))
	router.Post("/api/shorten/batch", Middlewares(router.PostAPIShortenBatchHandler, middlewares))

	return router
}

func (h *Handler) POSTHandler(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	idCookie, err := r.Cookie("user_id")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	url := string(b)
	id, err := h.Storage.Set(r.Context(), storage.CreateURL{
		URL:  url,
		User: idCookie.Value,
	})

	if err != nil {
		var pge *pgconn.PgError
		if errors.As(err, &pge) && pge.Code == pgerrcode.UniqueViolation {
			url, err := h.Storage.GetByOriginURL(r.Context(), url)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			resultURL := h.BaseURL + "/" + fmt.Sprintf("%d", url.ID)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(resultURL))
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resultURL := h.BaseURL + "/" + fmt.Sprintf("%d", id)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(201)
	w.Write([]byte(resultURL))
}

func (h *Handler) GETHandler(w http.ResponseWriter, r *http.Request) {
	rawID := chi.URLParam(r, "id")
	id, err := strconv.Atoi(rawID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	createURL, err := h.Storage.Get(r.Context(), id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	http.Redirect(w, r, createURL.URL, http.StatusTemporaryRedirect)

}

func (h *Handler) JSONHandler(w http.ResponseWriter, r *http.Request) {
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

	idCookie, err := r.Cookie("user_id")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := h.Storage.Set(r.Context(), storage.CreateURL{
		URL:  request.URL,
		User: idCookie.Value,
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

func (h *Handler) GetUrlsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idCookie, err := r.Cookie("user_id")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	createURLs, err := h.Storage.GetByUser(r.Context(), idCookie.Value)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(createURLs) == 0 {
		http.Error(w, "Not found", http.StatusNoContent)
		return
	}

	var shortenUrls []ShortURLs

	for _, val := range createURLs {
		shortenUrls = append(shortenUrls, ShortURLs{
			ShortURL:    h.BaseURL + "/" + fmt.Sprintf("%d", val.ID),
			OriginalURL: val.URL,
		})
	}

	res, err := json.Marshal(shortenUrls)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func (h *Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.Storage.Ping(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
func (h *Handler) PostAPIShortenBatchHandler(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var batchRequests []BatchRequest
	if err := json.Unmarshal(b, &batchRequests); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	idCookie, err := r.Cookie("user_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var ShortenBatch []storage.ShortenBatch
	for _, batchRequest := range batchRequests {
		ShortenBatch = append(ShortenBatch, storage.ShortenBatch{
			User:          idCookie.Value,
			URL:           batchRequest.OriginURL,
			CorrelationID: batchRequest.CorrelationID,
		})
	}

	ShortenBatchs, err := h.Storage.APIShortenBatch(r.Context(), ShortenBatch)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var batchResponses []BatchResponse
	for _, batchResponse := range ShortenBatchs {
		batchResponses = append(batchResponses, BatchResponse{
			CorrelationID: batchResponse.CorrelationID,
			ShortURL:      h.BaseURL + "/" + fmt.Sprintf("%d", batchResponse.ID),
		})
	}

	res, err := json.Marshal(batchResponses)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(res)
}

func (h *Handler) formatResult(id int) ([]byte, error) {
	resultURL := h.BaseURL + "/" + fmt.Sprintf("%d", id)
	response := Response{Result: resultURL}
	return json.Marshal(response)
}
