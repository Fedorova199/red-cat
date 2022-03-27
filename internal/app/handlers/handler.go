package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Fedorova199/red-cat/internal/app/middlewares"
	"github.com/Fedorova199/red-cat/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	*chi.Mux
	Storage storage.Storage
	BaseURL string
}

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

func NewHandler(storage storage.Storage, baseURL string) *Handler {
	router := &Handler{
		Mux:     chi.NewMux(),
		Storage: storage,
		BaseURL: baseURL,
	}
	router.Use(middlewares.GzipHandle)
	router.Use(middlewares.UngzipHandle)
	router.Post("/", router.POSTHandler)
	router.Post("/api/shorten", router.JSONHandler)
	router.Get("/{id}", router.GETHandler)

	return router
}

func (h *Handler) POSTHandler(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	url := string(b)
	id, err := h.Storage.Set(url)

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

	originURL, err := h.Storage.Get(id)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	http.Redirect(w, r, originURL, http.StatusTemporaryRedirect)

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

	id, err := h.Storage.Set(request.URL)

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
