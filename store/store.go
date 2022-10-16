package store

import (
	"context"

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
	c, err := pgx.Connect(context.Background(), `postgres://postgres:postgrespw@localhost:49153/testing`)
	if err != nil {
		panic(err)
	}

	user.Migration(c)

	return New(context.Background(), c)
}()
