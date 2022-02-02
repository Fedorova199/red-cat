package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/Fedorova199/red-cat/internal/config"
	"github.com/Fedorova199/red-cat/internal/server"
	"github.com/Fedorova199/red-cat/internal/services"
	"github.com/Fedorova199/red-cat/internal/storage"
)

func main() {
	conf, err := config.NewFromEnvAndCMD()
	if err != nil {
		panic(err)
	}
	persistentStorage, err := storage.NewFileStorage(conf.FileStoragePath)
	if err != nil {
		panic(err)
	}
	defer persistentStorage.Close()
	storageVar, err := storage.New(persistentStorage)
	if err != nil {
		panic(err)
	}
	serviceVar := services.New(storageVar)
	serverVar, err := server.New(
		serviceVar,
		server.WithAddress(conf.ServerAddress),
		server.WithURL(conf.BaseURL),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("Started at", conf.ServerAddress)
	go serverVar.ListenAndServe()

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Waiting for SIGINT (kill -2)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := serverVar.Shutdown(ctx); err != nil {
		panic("unexpected err on graceful shutdown")
	}
	fmt.Println("main: done. exiting")
}
