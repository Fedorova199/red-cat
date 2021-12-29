package main

import (
	"net/http"

	"github.com/Fedorova199/red-cat/cmd/shortener/handlers"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", handlers.BodyHandler)
	router.HandleFunc("/{id}", handlers.QueryHandler)
	http.Handle("/", router)
	// запуск сервера с адресом localhost, порт 8080
	http.ListenAndServe(":8080", nil)
}
