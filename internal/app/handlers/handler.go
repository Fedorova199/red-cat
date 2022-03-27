package handlers

import (
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
func NewHandler(storage storage.Storage, baseURL string, middlewares []Middleware) *Handler {
	router := &Handler{
		Mux:     chi.NewMux(),
		Storage: storage,
		BaseURL: baseURL,
	}
	router.Get("/{id}", Middlewares(router.GETHandler, middlewares))
	router.Get("/api/user/urls", Middlewares(router.GetUrlsHandler, middlewares))
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
	id, err := h.Storage.Set(idCookie.Value, url)

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

	createURL, err := h.Storage.Get(id)

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

	id, err := h.Storage.Set(idCookie.Value, request.URL)

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
	idCookie, err := r.Cookie("user_id")
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	createURLs, err := h.Storage.GetByUser(idCookie.Value)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
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
