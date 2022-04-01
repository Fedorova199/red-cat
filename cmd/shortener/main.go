package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Fedorova199/red-cat/internal/app/config"
	"github.com/Fedorova199/red-cat/internal/app/handlers"
	"github.com/Fedorova199/red-cat/internal/app/middlewares"
	"github.com/Fedorova199/red-cat/internal/app/storage"
	"github.com/Fedorova199/red-cat/internal/interfaces"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalln(err)
	}
	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	storage, err := storage.CreateDatabase(db)
	if err != nil {
		log.Fatalln(err)
	}

	ms := []interfaces.Middleware{
		middlewares.GzipEncoder{},
		middlewares.GzipDecoder{},
		middlewares.NewAuth([]byte("secret key")),
	}

	handler := handlers.NewHandler(storage, cfg.BaseURL, ms)
	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: handler,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		<-c
		server.Close()
	}()

	log.Fatal(server.ListenAndServe())
}
