package internal

import (
	"database/sql"
	"testing"

	sqil "github.com/hyphengolang/prelude/sql"
	"github.com/hyphengolang/prelude/testing/is"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestUser(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	t.Run(`pgx with database/sql`, func(t *testing.T) {
		connString := `postgres://postgres:postgrespw@localhost:49153/testing`
		db, err := sql.Open("pgx", connString)
		is.NoErr(err) // connect to database

		_, err = db.Exec(`
		begin;

		drop type account_type;

		create type account_type as enum ('guest', 'registered', 'admin');
		create temp table "my_test" (
			value account_type
		);

		insert into my_test values ('guest');
		insert into my_test values ('admin');

		commit;
				`)
		is.NoErr(err) // creating a dummy table

		// r := db.QueryRow(`select * from my_test where value = $1`, AdminUser)

		// err = r.Scan(&typ)

		var typ UserTyp

		err = sqil.QueryRow(db, func(r *sql.Row) error { return r.Scan(&typ) }, `select * from my_test where value = @value`, sql.Named("value", AdminUser))
		is.NoErr(err) // select from database
		is.Equal(typ, AdminUser)
	})

	// 	t.Run(`encoding a user interface`, func(t *testing.T) {
	// 		user := User{
	// 			ID:       suid.NewUUID(),
	// 			Username: "Fizz",
	// 			Email:    email.MustParse("fizz@mail.com"),
	// 			Password: password.MustParse("this_i$_f1zZ").MustHash(),
	// 		}

	// 		is.Equal(user.Username, "Fizz")
	// 	})

	// 	connString := `postgres://postgres:postgrespw@localhost:49153/testing`

	// 	t.Run(`using custom enum with SQL`, func(t *testing.T) {

	// 		conn, _ := pgx.Connect(context.Background(), connString)
	// 		t.Cleanup(func() {
	// 			conn.Close(context.Background())
	// 		})

	// 		_, err := conn.Exec(context.Background(), `
	// begin;

	// drop type account_type;

	// create type account_type as enum ('guest', 'registered', 'admin');
	// create temp table "my_test" (
	// 	value account_type
	// );

	// insert into my_test values ('guest');
	// insert into my_test values ('admin');

	// commit;
	// 		`)
	// 		is.NoErr(err) // creating a dummy table

	// 		row := conn.QueryRow(context.Background(), `select * from my_test where value = $1`, Guest)
	// 		var typ UserTyp
	// 		err = row.Scan(&typ)
	// 		is.NoErr(err) // select from database

	//		// err = psql.Exec(conn, `insert into my_test values ($1);`, Registered)
	//		// is.NoErr(err) // insert a new entry with enum value
	//	})
}

// https://dev.to/yogski/dealing-with-enum-type-in-postgresql-1j3g
