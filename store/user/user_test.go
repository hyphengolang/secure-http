package user

import (
	"context"
	"testing"

	"github.com/hyphengolang/prelude/testing/is"
	"github.com/jackc/pgx/v5"
	"secure.adoublef.com/internal"
	"secure.adoublef.com/internal/suid"
)

var r internal.UserRepo

func init() {
	connString := `postgres://postgres:postgrespw@localhost:49153/testing`

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

	r = NewRepo(context.Background(), c)
}

func TestRepo(t *testing.T) {
	t.Parallel()
	is, ctx := is.New(t), context.TODO()

	t.Cleanup(func() { r.Close(ctx) })

	t.Run(`select many from "account"`, func(t *testing.T) {
		as, err := r.SelectMany(ctx)
		is.NoErr(err)        // cannot query from database
		is.Equal(len(as), 0) // no users in database
	})

	fizzId := suid.NewUUID()
	buzzEmail := internal.Email("buzz@mail.com")
	burpUsername := "i_am_burp"
	t.Run(`insert one into "account"`, func(t *testing.T) {
		fizz := internal.User{
			ID:       fizzId,
			Username: "i_am_fizz",
			Email:    "fizz@mail.com",
			Password: internal.Password("p4$$w4rD").MustHash(),
		}

		err := r.Insert(ctx, &fizz)
		is.NoErr(err) // inserting "fizz"

		buzz := internal.User{
			ID:       suid.NewUUID(),
			Username: "i_am_buzz",
			Email:    buzzEmail,
			Password: internal.Password("p4$$w4rD").MustHash(),
		}

		err = r.Insert(ctx, &buzz)
		is.NoErr(err) // inserting "buzz"

		burp := internal.User{
			ID:       suid.NewUUID(),
			Username: burpUsername,
			Email:    "burp@mail.com",
			Password: internal.Password("p4$$w4rD").MustHash(),
		}

		err = r.Insert(ctx, &burp)
		is.NoErr(err) // inserting "buzz"

		us, err := r.SelectMany(ctx)
		is.NoErr(err)        // cannot query from database
		is.Equal(len(us), 3) // 2 users in database
	})

	t.Run(`invalid inserts`, func(t *testing.T) {
		u := internal.User{
			ID:       suid.NewUUID(),
			Username: "i_am_fizz",
			Email:    "fizz@mail.com",
			Password: internal.Password("p4$$w4rD").MustHash(),
		}

		err := r.Insert(context.Background(), &u)
		is.True(err != nil) // "fizz" already exists

		u = internal.User{
			ID:       suid.NewUUID(),
			Username: "i_am_bazz",
			Email:    "bazzmail.com",
			Password: internal.Password("p4$$w4rD").MustHash(),
		}

		err = r.Insert(context.Background(), &u)
		is.True(err != nil) // invalid email

		us, err := r.SelectMany(ctx)
		is.NoErr(err)        // cannot query from database
		is.Equal(len(us), 3) // 3 users in database
	})

	t.Run(`delete one from "account"`, func(t *testing.T) {
		err := r.Delete(ctx, fizzId)
		is.NoErr(err) // perform soft delete on "i_am_fizz"

		us, _ := r.SelectMany(ctx)
		is.Equal(len(us), 3) // 3 users in database

		ctx = context.WithValue(ctx, RuleSoftDeletion, HardDelete)
		err = r.Delete(ctx, buzzEmail)
		is.NoErr(err) // perform hard delete on "i_am_buzz"

		us, _ = r.SelectMany(ctx)
		is.Equal(len(us), 2) // 2 users in database

		err = r.Delete(ctx, nil)
		is.True(err != nil) // invalid key
	})

	t.Run(`select one from "account"`, func(t *testing.T) {
		u, err := r.Select(ctx, fizzId)
		is.NoErr(err)                            // select "i_am_fizz"
		is.Equal(u.ID.String(), fizzId.String()) // "id" values are the same

		u, err = r.Select(ctx, burpUsername)
		is.NoErr(err)                      // select "i_am_burp"
		is.Equal(u.Username, burpUsername) // "username" values are the same
	})
}
