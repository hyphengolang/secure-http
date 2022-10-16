package user

import (
	"context"
	"errors"
	"time"

	"github.com/hyphengolang/prelude/types/email"
	"github.com/hyphengolang/prelude/types/password"
	"github.com/hyphengolang/prelude/types/suid"
	"github.com/jackc/pgx/v5"
	"secure.adoublef.com/internal"
	"secure.adoublef.com/internal/psql"
)

type User struct {
	ID        suid.UUID
	Username  string
	Email     email.Email
	Password  password.PasswordHash
	CreatedAt time.Time
	IsDeleted bool
	DeletedAt *time.Time
}

const (
	qrySelectMany = `select id, username, email, password from "account"`

	qrySelectByID       = `select id, username, email, password from "account" where id = $1`
	qrySelectByEmail    = `select id, username, email, password from "account" where email = $1`
	qrySelectByUsername = `select id, username, email, password from "account" where username = $1`

	qryInsert = `insert into "account" (id, username, email, password) values (@id, @username, @email, @password)`

	qryDeleteByID    = `delete from "account" where id = $1;`
	qryDeleteByEmail = `delete from "account" where email = $1;`

	setRuleSoftDeletionOn  = `set rules.soft_deletion to 'on'`
	setRuleSoftDeletionOff = `set rules.soft_deletion to 'off'`
)

func (r Repo) Select(ctx context.Context, key any) (*internal.User, error) {
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
	return &u, psql.QueryRow(r.q, qry, func(r pgx.Row) error { return r.Scan(&u.ID, &u.Username, &u.Email, &u.Password) }, key)
}

func (r Repo) SelectMany(ctx context.Context) ([]internal.User, error) {
	return psql.Query(r.q, qrySelectMany, func(r pgx.Rows, u *internal.User) error { return r.Scan(&u.ID, &u.Username, &u.Email, &u.Password) })
}

func (r Repo) Insert(ctx context.Context, u *internal.User) error {
	args := pgx.NamedArgs{
		"id":       u.ID,
		"username": u.Username,
		"email":    u.Email,
		"password": u.Password,
	}

	return psql.Exec(r.q, qryInsert, args)
}

func (r Repo) Delete(ctx context.Context, key any) error {
	tx, err := r.q.Begin(ctx)
	if err != nil {
		return err
	}

	var rule string
	if typ, set := ctx.Value(RuleSoftDeletion).(DeleteTyp); !set {
		rule = setRuleSoftDeletionOn
	} else {
		switch typ {
		case HardDelete:
			rule = setRuleSoftDeletionOff
		default:
			rule = setRuleSoftDeletionOn
		}
	}

	if err = psql.Exec(tx, rule); err != nil {
		return err
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

	if err = psql.Exec(tx, qry, key); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

type Repo struct {
	ctx context.Context

	q *pgx.Conn
}

func (r Repo) Context() context.Context {
	if r.ctx != nil {
		return r.ctx
	}
	return context.Background()
}

func (r *Repo) Close(ctx context.Context) error { return r.q.Close(ctx) }

func NewRepo(ctx context.Context, q *pgx.Conn) internal.UserRepo {
	r := &Repo{ctx, q}
	return r
}

var ErrInvalidType = errors.New(`invalid type`)

// More info regarding soft deleting https://evilmartians.com/chronicles/soft-deletion-with-postgresql-but-with-logic-on-the-database
type DeleteTyp int

func (t DeleteTyp) String() string {
	return [...]string{
		"soft_delete",
		"hard_delete",
	}[t]
}

const (
	SoftDelete DeleteTyp = iota
	HardDelete
)

type contextKey string

const (
	RuleSoftDeletion = contextKey("rule-soft-deletion")
)

var RepoTest = func() internal.UserRepo {
	ctx := context.Background()
	c, err := pgx.Connect(ctx, `postgres://postgres:postgrespw@localhost:49153/testing`)
	if err != nil {
		panic(err)
	}

	Migration(c)

	return NewRepo(ctx, c)
}()

// For development only
//
// If there is an error, it will panic immediately
func Migration(c *pgx.Conn) {
	if _, err := c.Exec(context.Background(), migration); err != nil {
		panic(err)
	}
}

const migration = `
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
`
