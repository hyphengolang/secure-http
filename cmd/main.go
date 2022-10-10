package main

import (
	"context"
	"log"
	"net/http"

	"secure.adoublef.com/service"
	"secure.adoublef.com/store"
)

func dev() error {
	store := store.Dev(context.Background(), `postgres://postgres:postgrespw@localhost:49153/testing`)
	handler := service.New(context.Background(), store)

	srv := http.Server{
		Addr:    ":7878",
		Handler: handler,
	}

	return srv.ListenAndServe()
}

func main() {
	if err := dev(); err != nil {
		log.Fatalln(err)
	}
}
