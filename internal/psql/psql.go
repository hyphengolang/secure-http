package psql

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func Query[T any](db *pgx.Conn, query string, scanner func(r pgx.Rows, v *T) error, args ...any) ([]T, error) {
	return QueryContext(context.Background(), db, query, scanner, args...)
}

func QueryRow[Q *pgx.Conn](db Q, query string, scanner func(r pgx.Row) error, args ...any) error {
	return QueryRowContext(context.Background(), db, query, scanner, args...)
}

func Exec(db *pgx.Conn, query string, args ...any) error {

	return ExecContext(context.Background(), db, query, args...)
}

func QueryContext[T any](ctx context.Context, db *pgx.Conn, query string, scanner func(r pgx.Rows, v *T) error, args ...any) ([]T, error) {
	rows, err := db.Query(ctx, query, args...)
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

func QueryRowContext(ctx context.Context, db *pgx.Conn, query string, scanner func(r pgx.Row) error, args ...any) error {
	return scanner(db.QueryRow(ctx, query, args...))
}

func ExecContext(ctx context.Context, db *pgx.Conn, query string, args ...any) error {
	_, err := db.Exec(ctx, query, args...)
	return err
}
