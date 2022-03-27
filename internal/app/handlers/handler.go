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

// type Storage interface {
// 	Get(ctx context.Context, id uint64) (storages.Record, error)
// 	GetByOriginURL(ctx context.Context, originURL string) (storages.Record, error)
// 	GetByUser(ctx context.Context, userID string) ([]storages.Record, error)
// 	Put(ctx context.Context, record storages.Record) (uint64, error)
// 	PutRecords(ctx context.Context, records []storages.BatchRecord) ([]storages.BatchRecord, error)
// 	Ping(ctx context.Context) error
// }

// type Middleware interface {
// 	Handle(next http.HandlerFunc) http.HandlerFunc
// }

func (h *Handler) PostHandler(w http.ResponseWriter, r *http.Request) {
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
		User: idCookie.Value,
		URL:  url,
	})

	if err != nil {
		var pge *pgconn.PgError
		if errors.As(err, &pge) && pge.Code == pgerrcode.UniqueViolation {
			createURL, err := h.Storage.GetOriginURL(r.Context(), url)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			resultURL := h.BaseURL + "/" + fmt.Sprintf("%d", createURL.ID)
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
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resultURL))
}

func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	rawID := chi.URLParam(r, "id")
	id, err := strconv.Atoi(rawID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	origin, err := h.Storage.Get(r.Context(), id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	http.Redirect(w, r, origin.URL, http.StatusTemporaryRedirect)
}

func (h *Handler) JSONHandler(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	request := storage.Request{}
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
		User: idCookie.Value,
		URL:  request.URL,
	})

	if err != nil {
		var pge *pgconn.PgError
		if errors.As(err, &pge) && pge.Code == pgerrcode.UniqueViolation {
			createURL, err := h.Storage.GetOriginURL(r.Context(), request.URL)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			res, err := h.formatResult(createURL.ID)
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

func (h *Handler) formatResult(id int) ([]byte, error) {
	resultURL := h.BaseURL + "/" + fmt.Sprintf("%d", id)
	response := storage.Response{Result: resultURL}
	return json.Marshal(response)
}

func (h *Handler) GetUrlsHandler(w http.ResponseWriter, r *http.Request) {
	idCookie, err := r.Cookie("user_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	getuser, err := h.Storage.GetUser(r.Context(), idCookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(getuser) == 0 {
		http.Error(w, "Not found", http.StatusNoContent)
		return
	}

	var shortenUrls []storage.ShortURLs
	for _, shortURL := range getuser {
		shortenUrls = append(shortenUrls, storage.ShortURLs{
			ShortURL:    h.BaseURL + "/" + fmt.Sprintf("%d", shortURL.ID),
			OriginalURL: shortURL.URL,
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var batchRequests []storage.BatchRequest
	if err := json.Unmarshal(b, &batchRequests); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	idCookie, err := r.Cookie("user_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var shortBatch []storage.ShortenBatch
	for _, batchRequest := range batchRequests {
		shortBatch = append(shortBatch, storage.ShortenBatch{
			User:          idCookie.Value,
			URL:           batchRequest.OriginURL,
			CorrelationID: batchRequest.CorrelationID,
		})
	}

	shortBatch, err = h.Storage.PutBatch(r.Context(), shortBatch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var batchResponses []storage.BatchResponse
	for _, batchresp := range shortBatch {
		batchResponses = append(batchResponses, storage.BatchResponse{
			CorrelationID: batchresp.CorrelationID,
			ShortURL:      h.BaseURL + "/" + fmt.Sprintf("%d", batchresp.ID),
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
