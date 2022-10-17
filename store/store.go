package store

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
	"secure.adoublef.com/internal"
	"secure.adoublef.com/store/user"
)

type Store struct {
	u internal.UserRepo
}

func (s Store) UserRepo() internal.UserRepo { return s.u }

func New(ctx context.Context, c *pgx.Conn) *Store {
	return &Store{
		u: user.NewRepo(ctx, c),
	}
}

var StoreTest = func() *Store {
	ctx := context.Background()

	connString := os.ExpandEnv("host=${POSTGRES_HOSTNAME} port=${DB_PORT} user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DB} sslmode=${SSL_MODE}")
	c, err := pgx.Connect(ctx, connString)
	if err != nil {
		panic(err)
	}

	user.Migration(c)

	return New(ctx, c)
}()
