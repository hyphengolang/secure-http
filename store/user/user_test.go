package user

import (
	"context"
	"testing"

	"github.com/hyphengolang/prelude/testing/is"
	"github.com/hyphengolang/prelude/types/email"
	"github.com/hyphengolang/prelude/types/password"
	"github.com/hyphengolang/prelude/types/suid"
	"github.com/jackc/pgx/v5/pgxpool"
	"secure.adoublef.com/internal"
)

var testRepo *Repo

func init() {
	connString := `postgres://postgres:postgrespw@localhost:49153/testing`
	p, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		panic(err)
	}

	Migration(p)

	testRepo = NewRepo(context.Background(), p)
}

func TestRepo(t *testing.T) {
	t.Parallel()
	is, ctx := is.New(t), context.TODO()

	t.Cleanup(func() { testRepo.p.Close() })

	t.Run(`select many from "account"`, func(t *testing.T) {
		as, err := testRepo.SelectMany(ctx)
		is.NoErr(err)        // cannot query from database
		is.Equal(len(as), 0) // no users in database
	})

	fizzId := suid.NewUUID()
	buzzEmail := email.Email("buzz@mail.com")
	burpUsername := "i_am_burp"
	t.Run(`insert one into "account"`, func(t *testing.T) {
		fizz := internal.User{
			ID:       fizzId,
			Username: "i_am_fizz",
			Email:    "fizz@mail.com",
			Password: password.Password("p4$$w4rD").MustHash(),
		}

		err := testRepo.Insert(ctx, &fizz)
		is.NoErr(err) // inserting "fizz"

		buzz := internal.User{
			ID:       suid.NewUUID(),
			Username: "i_am_buzz",
			Email:    buzzEmail,
			Password: password.Password("p4$$w4rD").MustHash(),
		}

		err = testRepo.Insert(ctx, &buzz)
		is.NoErr(err) // inserting "buzz"

		burp := internal.User{
			ID:       suid.NewUUID(),
			Username: burpUsername,
			Email:    "burp@mail.com",
			Password: password.Password("p4$$w4rD").MustHash(),
		}

		err = testRepo.Insert(ctx, &burp)
		is.NoErr(err) // inserting "buzz"

		us, err := testRepo.SelectMany(ctx)
		is.NoErr(err)        // cannot query from database
		is.Equal(len(us), 3) // 2 users in database
	})

	t.Run(`invalid inserts`, func(t *testing.T) {
		u := internal.User{
			ID:       suid.NewUUID(),
			Username: "i_am_fizz",
			Email:    "fizz@mail.com",
			Password: password.Password("p4$$w4rD").MustHash(),
		}

		err := testRepo.Insert(context.Background(), &u)
		is.True(err != nil) // "fizz" already exists

		u = internal.User{
			ID:       suid.NewUUID(),
			Username: "i_am_bazz",
			Email:    "bazzmail.com",
			Password: password.Password("p4$$w4rD").MustHash(),
		}

		err = testRepo.Insert(context.Background(), &u)
		is.True(err != nil) // invalid email

		us, err := testRepo.SelectMany(ctx)
		is.NoErr(err)        // cannot query from database
		is.Equal(len(us), 3) // 3 users in database
	})

	t.Run(`select one from "account"`, func(t *testing.T) {
		_, err := testRepo.Select(ctx, fizzId)
		is.NoErr(err) // account "i_am_fizz" does not exist

		u, err := testRepo.Select(ctx, burpUsername)
		is.NoErr(err)                      // select "i_am_burp"
		is.Equal(u.Username, burpUsername) // "username" values are the same
	})

	t.Run(`delete one from "account"`, func(t *testing.T) {
		err := testRepo.Delete(ctx, fizzId)
		is.NoErr(err) // delete account "i_am_fizz"

		us, _ := testRepo.SelectMany(ctx)
		is.Equal(len(us), 2) // 2 users in database

		err = testRepo.Delete(ctx, nil)
		is.True(err != nil) // invalid key
	})
}
