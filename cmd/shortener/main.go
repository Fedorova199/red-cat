package main

import (
	"database/sql"
	"flag"
	"github.com/Fedorova199/red-cat/internal/app/handlers"
	"github.com/Fedorova199/red-cat/internal/app/middlewares"
	"github.com/Fedorova199/red-cat/internal/app/storages"
	"github.com/caarlos0/env/v6"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

func main() {
	cfg := parseVariables()

	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	storage, err := storages.CreateDatabaseStorage(db)
	if err != nil {
		log.Fatal(err)
	}

	mws := []handlers.Middleware{
		middlewares.GzipEncoder{},
		middlewares.GzipDecoder{},
		middlewares.NewAuthenticator([]byte("secret key")),
	}

	handler := handlers.NewHandler(storage, cfg.BaseURL, mws)
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

func parseVariables() Config {
	var cfg = Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "db.txt",
	}

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Server address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "File storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database DSN")
	flag.Parse()

	return cfg
}
