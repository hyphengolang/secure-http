package user

import (
	"context"
	"errors"
	"time"

	psql "github.com/hyphengolang/prelude/sql/postgres"
	"github.com/hyphengolang/prelude/types/email"
	"github.com/hyphengolang/prelude/types/password"
	"github.com/hyphengolang/prelude/types/suid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"secure.adoublef.com/internal"
)

type User struct {
	ID        suid.UUID
	Username  string
	Email     email.Email
	Password  password.PasswordHash
	CreatedAt time.Time
	DeletedAt *time.Time
}

const (
	migration = `
	begin;
	
	create extension if not exists "uuid-ossp";
	create extension if not exists "citext";
	
	create temp table if not exists "user" (
		id uuid primary key default uuid_generate_v4(),
		username text unique not null check (username <> ''),
		email citext unique not null check (email ~ '^[a-zA-Z0-9.!#$%&''*+/=?^_{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$'),
		password citext not null check (password <> ''),
		created_at timestamp not null default now()
	);
	
	commit;
	`
	qrySelectMany = `select id, username, email, password from "user"`

	qrySelectByID       = `select id, username, email, password from "user" where id = $1`
	qrySelectByEmail    = `select id, username, email, password from "user" where email = $1`
	qrySelectByUsername = `select id, username, email, password from "user" where username = $1`

	qryInsert = `insert into "user" (id, username, email, password) values (@id, @username, @email, @password)`

	qryDeleteByID    = `delete from "user" where id = $1;`
	qryDeleteByEmail = `delete from "user" where email = $1;`
)

func (r *Repo) Select(ctx context.Context, key any) (*internal.User, error) {
	var qry string
	switch key.(type) {
	case suid.UUID:
		qry = qrySelectByID
	case email.Email:
		qry = qrySelectByEmail
	case string:
		qry = qrySelectByUsername
	default:
		return nil, ErrInvalidType
	}
	var u internal.User
	return &u, psql.QueryRowContext(ctx, r.p, qry, func(r pgx.Row) error { return r.Scan(&u.ID, &u.Username, &u.Email, &u.Password) }, key)
}

func (r *Repo) SelectMany(ctx context.Context) ([]internal.User, error) {
	return psql.QueryContext(ctx, r.p, qrySelectMany, func(r pgx.Rows, u *internal.User) error { return r.Scan(&u.ID, &u.Username, &u.Email, &u.Password) })
}

func (r *Repo) Insert(ctx context.Context, u *internal.User) error {
	args := pgx.NamedArgs{
		"id":       u.ID,
		"username": u.Username,
		"email":    u.Email,
		"password": u.Password,
	}

	return psql.ExecContext(ctx, r.p, qryInsert, args)
}

func (r *Repo) Delete(ctx context.Context, key any) error {
	// return deleteNotWorking(ctx, r.p, key)
	return deleteWorking(ctx, r.p, key)
}

func deleteWorking(ctx context.Context, p *pgxpool.Pool, key any) error {
	var qry string
	switch key.(type) {
	case suid.UUID:
		qry = qryDeleteByID
	case email.Email:
		qry = qryDeleteByEmail
	default:
		return ErrInvalidType
	}

	return psql.ExecContext(ctx, p, qry, key)
}

// TimeOut Error
func deleteNotWorking(ctx context.Context, p *pgxpool.Pool, key any) error {
	tx, err := p.Begin(ctx)
	if err != nil {
		panic(err)
	}

	var qry string
	switch key.(type) {
	case suid.UUID:
		qry = qryDeleteByID
	case email.Email:
		qry = qryDeleteByEmail
	default:
		return ErrInvalidType
	}

	if _, err := tx.Exec(ctx, qry, key); err != nil {
		panic(err)
	}

	if err := tx.Commit(ctx); err != nil {
		panic(err)
	}

	return nil
}

type Repo struct {
	ctx context.Context

	p *pgxpool.Pool
}

func (r *Repo) Context() context.Context {
	if r.ctx != nil {
		return r.ctx
	}
	return context.Background()
}

func (r *Repo) Close() { r.p.Close() }

func NewRepo(ctx context.Context, q *pgxpool.Pool) *Repo {
	r := &Repo{ctx, q}
	return r
}

var ErrInvalidType = errors.New(`invalid type`)

var RepoTest = func() internal.UserRepo {
	ctx := context.Background()

	connString := `postgres://postgres:postgrespw@localhost:49153/testing`
	p, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		panic(err)
	}

	Migration(p)

	return NewRepo(ctx, p)
}()

// For development only
//
// If there is an error, it will panic immediately
func Migration(p *pgxpool.Pool) {
	if _, err := p.Exec(context.Background(), migration); err != nil {
		panic(err)
	}
}
