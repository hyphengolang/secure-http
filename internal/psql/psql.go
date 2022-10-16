package psql

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Q interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func Query[T any](q Q, query string, scanner func(r pgx.Rows, v *T) error, args ...any) ([]T, error) {
	return QueryContext(context.Background(), q, query, scanner, args...)
}

func QueryRow(q Q, query string, scanner func(r pgx.Row) error, args ...any) error {
	return QueryRowContext(context.Background(), q, query, scanner, args...)
}

func Exec(q Q, query string, args ...any) error {

	return ExecContext(context.Background(), q, query, args...)
}

func QueryContext[T any](ctx context.Context, q Q, query string, scanner func(r pgx.Rows, v *T) error, args ...any) ([]T, error) {
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var vs []T
	for rows.Next() {
		var v T
		err = scanner(rows, &v)
		if err != nil {
			return nil, err
		}
		vs = append(vs, v)
	}
	return vs, rows.Err()
}

func QueryRowContext(ctx context.Context, q Q, query string, scanner func(r pgx.Row) error, args ...any) error {
	return scanner(q.QueryRow(ctx, query, args...))
}

func ExecContext(ctx context.Context, q Q, query string, args ...any) error {
	_, err := q.Exec(ctx, query, args...)
	return err
}
