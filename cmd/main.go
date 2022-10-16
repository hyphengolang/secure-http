package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"secure.adoublef.com/file"
)

func dev() error {
	// handler := service.New(context.Background(), store.StoreTest)

	handler := file.New(chi.NewMux())

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
