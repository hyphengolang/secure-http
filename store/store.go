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

func StoreTest(ctx context.Context, connString string) *Store {
	c, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		panic(err)
	}

	_, err = c.Exec(context.Background(), `
	begin;

		create extension if not exists "uuid-ossp";
		create extension if not exists "citext";

		create temp table if not exists "account" (
			id uuid primary key default uuid_generate_v4(),
			username text unique not null check (username <> ''),
			email citext unique not null check (email ~ '^[a-zA-Z0-9.!#$%&''*+/=?^_{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$'),
			password citext not null check (password <> ''),
			created_at timestamp not null default now(),
			deleted boolean not null default false
		);
		
		create or replace rule "_soft_deletion" 
			as on delete to "account" 
			where current_setting('rules.soft_deletion') = 'on'
			do instead update "account" set deleted = true where id = old.id;

		set rules.soft_deletion to 'on';
		
	commit;
	`)
	if err != nil {
		panic(err)
	}

	return New(ctx, c)
}
