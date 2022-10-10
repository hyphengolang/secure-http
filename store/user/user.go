package user

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"secure.adoublef.com/internal"
	"secure.adoublef.com/internal/suid"
)

type User struct {
	ID        suid.UUID
	Username  string
	Email     internal.Email
	Password  internal.PasswordHash
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
	var row pgx.Row
	switch key := key.(type) {
	case suid.UUID:
		row = r.q.QueryRow(ctx, qrySelectByID, key)
	case internal.Email:
		row = r.q.QueryRow(ctx, qrySelectByEmail, key)
	case string:
		row = r.q.QueryRow(ctx, qrySelectByUsername, key)
	default:
		return nil, ErrInvalidType
	}

	var u internal.User
	return &u, row.Scan(&u.ID, &u.Username, &u.Email, &u.Password)
}

func (r Repo) SelectMany(ctx context.Context) ([]internal.User, error) {
	rs, err := r.q.Query(ctx, qrySelectMany)
	if err != nil {
		return nil, err
	}
	defer rs.Close()

	var us []internal.User
	for rs.Next() {
		var u User
		_ = rs.Scan(&u.ID, &u.Username, &u.Email, &u.Password)
		us = append(us, internal.User{
			ID:       u.ID,
			Email:    u.Email,
			Password: u.Password,
			Username: u.Username,
		})
	}

	return us, rs.Err()
}

func (r Repo) Insert(ctx context.Context, u *internal.User) error {
	args := pgx.NamedArgs{
		"id":       u.ID,
		"username": u.Username,
		"email":    u.Email,
		"password": u.Password,
	}

	_, err := r.q.Exec(ctx, qryInsert, args)
	return err
}

func (r Repo) Delete(ctx context.Context, key any) error {
	tx, err := r.q.Begin(ctx)
	if err != nil {
		return err
	}

	if typ, set := ctx.Value(RuleSoftDeletion).(DeleteTyp); !set {
		if _, err = tx.Exec(ctx, setRuleSoftDeletionOn); err != nil {
			return err
		}
	} else {
		switch typ {
		case HardDelete:
			if _, err = tx.Exec(ctx, setRuleSoftDeletionOff); err != nil {
				return err
			}
		default:
			if _, err = tx.Exec(ctx, setRuleSoftDeletionOn); err != nil {
				return err
			}
		}
	}

	switch key := key.(type) {
	case suid.UUID:
		if _, err = tx.Exec(ctx, qryDeleteByID, key); err != nil {
			return err
		}
	case internal.Email:
		if _, err = tx.Exec(ctx, qryDeleteByEmail, key); err != nil {
			return err
		}
	default:
		return ErrInvalidType
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

type DeleteTyp int

const (
	// More info regarding soft deleting https://evilmartians.com/chronicles/soft-deletion-with-postgresql-but-with-logic-on-the-database
	SoftDelete DeleteTyp = iota
	HardDelete
)

type contextKey string

const (
	RuleSoftDeletion = contextKey("rule-soft-deletion")
)
