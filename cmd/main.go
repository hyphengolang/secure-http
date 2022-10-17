package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"secure.adoublef.com/service"
	"secure.adoublef.com/store"
	"secure.adoublef.com/store/user"
)

var connString, srvAddr string

func init() {
	connString = os.ExpandEnv("host=${POSTGRES_HOSTNAME} port=${DB_PORT} user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DB} sslmode=${SSL_MODE}")
	srvAddr = os.ExpandEnv("${SERVER_HOSTNAME}:${SERVER_PORT}")
}

func dev() error {
	// setup store
	ctx := context.Background()

	c, err := pgx.Connect(ctx, connString)
	if err != nil {
		panic(err)
	}

	// will panic if error
	user.Migration(c)

	store := store.New(ctx, c)

	// connect to server
	handler := service.New(context.Background(), store)

	srv := http.Server{
		Addr:     srvAddr,
		Handler:  handler,
		ErrorLog: log.Default(),
	}

	srv.ErrorLog.Printf("now listening on %s", srv.Addr)

	return srv.ListenAndServe()
}

func main() {
	if err := dev(); err != nil {
		log.Fatalln(err)
	}
}
