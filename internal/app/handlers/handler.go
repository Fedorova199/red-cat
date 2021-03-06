package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"

	"github.com/Fedorova199/red-cat/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

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
		if errors.Is(storage.ErrDeleted, err) {
			http.Error(w, err.Error(), http.StatusGone)
			return
		}
	}

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

	createdURLs, err := h.Storage.GetUser(r.Context(), idCookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(createdURLs) == 0 {
		http.Error(w, "Not found", http.StatusNoContent)
		return
	}

	shortenUrls := make([]storage.ShortURLs, 0, len(createdURLs))

	for _, shortURL := range createdURLs {
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

	shortBatch := make([]storage.ShortenBatch, 0, len(batchRequests))
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

	batchResponses := make([]storage.BatchResponse, 0, len(shortBatch))

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

func (h *Handler) DeleteUrlsHandler(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var deleteIDs []string
	if err := json.Unmarshal(b, &deleteIDs); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ids := make([]int, 0, len(deleteIDs))
	for _, deleteID := range deleteIDs {
		id, err := strconv.Atoi(deleteID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ids = append(ids, id)
	}

	idCookie, err := r.Cookie("user_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	outCh := make(chan int)

	go func() {
		var idsToDelete []int
		for id := range outCh {
			idsToDelete = append(idsToDelete, id)
		}
		err = h.Storage.DeleteURLs(r.Context(), idsToDelete)
		if err != nil {
			return
		}
	}()

	wg := &sync.WaitGroup{}
	for _, id := range ids {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			record, err := h.Storage.Get(r.Context(), id)
			if err != nil {
				return
			}

			if record.User == idCookie.Value {
				outCh <- id
			}
		}(id)
	}

	w.WriteHeader(http.StatusAccepted)
	wg.Wait()
	close(outCh)
}
