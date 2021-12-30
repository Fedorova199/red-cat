package main

import (
	"net/http"

	"github.com/Fedorova199/red-cat/cmd/shortener/handlers"

	"github.com/gorilla/mux"
)

func main() {
	Url := handlers.NewUrl()
	router := mux.NewRouter()
	router.HandleFunc("/", Url.BodyHandler)
	router.HandleFunc("/{id}", Url.QueryHandler)
	http.Handle("/", router)
	// запуск сервера с адресом localhost, порт 8080
	http.ListenAndServe(":8080", nil)
}
