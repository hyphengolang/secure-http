package psql

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

type testcase struct {
	Id   int
	Name string
	Age  int
}

var psql = os.ExpandEnv("host=${POSTGRES_HOSTNAME} port=${DB_PORT} user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DB} sslmode=disable")

func TestQueryRow(t *testing.T) {
	db, _ := pgx.Connect(context.Background(), psql)

	var b testcase
	err := QueryRow(db, "SELECT * FROM testcase WHERE id = $1", func(row pgx.Row) error { return row.Scan(&b.Id, &b.Name, &b.Age) }, 1)
	assert.Assert(t, err)

	assert.Assert(t, cmp.Equal(b.Name, "John"))
}

func TestQuery(t *testing.T) {
	db, _ := pgx.Connect(context.Background(), psql)

	items, err := Query(db, "SELECT * FROM testcase WHERE age > $1", func(rows pgx.Rows, i *testcase) error {
		return rows.Scan(&i.Id, &i.Name, &i.Age)
	}, 24)

	assert.Assert(t, err)

	assert.Assert(t, cmp.Equal(len(items), 2))
	fmt.Printf("item: %v\n", items)
}

func TestExec(t *testing.T) {
	db, _ := pgx.Connect(context.Background(), psql)

	err := Exec(db, "INSERT INTO testcase (name,age) VALUES ($1,$2)", "Maxi", 31)
	assert.Assert(t, err)
}
