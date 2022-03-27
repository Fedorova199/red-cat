package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Fedorova199/red-cat/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	*chi.Mux
	Storage storage.Storage
	BaseURL string
	DB      *sql.DB
}

func NewHandler(storage storage.Storage, baseURL string, middlewares []Middleware, db *sql.DB) *Handler {
	router := &Handler{
		Mux:     chi.NewMux(),
		Storage: storage,
		BaseURL: baseURL,
		DB:      db,
	}
	router.Get("/{id}", Middlewares(router.GETHandler, middlewares))
	router.Get("/api/user/urls", Middlewares(router.GetUrlsHandler, middlewares))
	router.Get("/ping", Middlewares(router.PingHandler, middlewares))
	router.Post("/", Middlewares(router.POSTHandler, middlewares))
	router.Post("/api/shorten", Middlewares(router.JSONHandler, middlewares))

	return router
}

func (h *Handler) POSTHandler(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	idCookie, err := r.Cookie("user_id")

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	url := string(b)
	id, err := h.Storage.Set(r.Context(), storage.CreateURL{
		URL:  url,
		User: idCookie.Value,
	})

	if err != nil {
		http.Error(w, err.Error(), 500)
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
		http.Error(w, err.Error(), 500)
		return
	}

	createURL, err := h.Storage.Get(r.Context(), id)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	http.Redirect(w, r, createURL.URL, http.StatusTemporaryRedirect)

}

func (h *Handler) JSONHandler(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	request := Request{}
	if err := json.Unmarshal(b, &request); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	idCookie, err := r.Cookie("user_id")

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	id, err := h.Storage.Set(r.Context(), storage.CreateURL{
		URL:  request.URL,
		User: idCookie.Value,
	})

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	resultURL := h.BaseURL + "/" + fmt.Sprintf("%d", id)
	response := Response{Result: resultURL}

	res, err := json.Marshal(response)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(res)
}

func (h *Handler) GetUrlsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idCookie, err := r.Cookie("user_id")

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	createURLs, err := h.Storage.GetByUser(r.Context(), idCookie.Value)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if len(createURLs) == 0 {
		http.Error(w, "Not found", 204)
		return
	}

	shortenUrls := make([]ShortURLs, 0)

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

	w.WriteHeader(200)
	w.Write(res)
}

func (h *Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.DB.Ping(); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(200)
}
