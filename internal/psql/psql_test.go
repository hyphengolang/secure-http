package psql

import (
	"context"
	"os"
	"testing"

	"github.com/hyphengolang/prelude/testing/is"
	"github.com/jackc/pgx/v5"
)

var psql = os.ExpandEnv("host=${POSTGRES_HOSTNAME} port=${DB_PORT} user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DB} sslmode=disable")

func TestGenericQueries(t *testing.T) {
	t.Parallel()

	type testcase struct {
		Id   int
		Name string
		Age  int
	}

	db, _ := pgx.Connect(context.Background(), psql)
	is := is.New(t)

	t.Cleanup(func() { db.Close(context.Background()) })

	t.Run(`select entry from database`, func(t *testing.T) {
		var b testcase
		err := QueryRow(db, "SELECT * FROM testcase WHERE id = $1", func(row pgx.Row) error { return row.Scan(&b.Id, &b.Name, &b.Age) }, 1)
		is.NoErr(err) // can make query

		is.Equal(b.Name, "John") // user.Name == "John"
	})

	t.Run(`select all from database`, func(t *testing.T) {
		items, err := Query(db, "SELECT * FROM testcase WHERE age > $1", func(rows pgx.Rows, i *testcase) error {
			return rows.Scan(&i.Id, &i.Name, &i.Age)
		}, 24)

		is.NoErr(err) // can make query

		is.Equal(len(items), 2) // two entries in the test database
	})

	t.Run(`insert into database`, func(t *testing.T) {
		err := Exec(db, "INSERT INTO testcase (name,age) VALUES ($1,$2)", "Maxi", 31)
		is.NoErr(err) // no error making exec
	})

}
