package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/hyphengolang/prelude/testing/is"
	"github.com/jackc/pgx/v5"

	"secure.adoublef.com/store/user"
)

const (
	applicationJson = "application/json"
)

var h http.Handler

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

	r := user.NewRepo(context.Background(), c)
	h = NewService(context.Background(), chi.NewMux(), r)
}

func TestService(t *testing.T) {
	t.Parallel()
	is, ctx := is.New(t), context.TODO()
	_, _ = is, ctx

	srv := httptest.NewServer(h)

	t.Cleanup(func() { srv.Close() })

	fizzUrl := &url.URL{}
	t.Run("register some new accounts", func(t *testing.T) {
		payload := `
		{
			"username":"i_am_fizz",
			"email":"fizz@mail.com",
			"password":"p4$$w4rD"
		}`
		// v:=url.Values{}
		// v.Encode()
		res, _ := srv.Client().Post(srv.URL+"/api/v1/account/", applicationJson, strings.NewReader(payload))
		is.Equal(res.StatusCode, http.StatusCreated) // register "i_am_fizz"
		fizzUrl, _ = res.Location()

		payload = `
		{
			"username":"i_am_buzz",
			"email":"buzz@mail.com",
			"password":"p4$$w4rD"
		}`

		res, _ = srv.Client().Post(srv.URL+"/api/v1/account/", applicationJson, strings.NewReader(payload))
		is.Equal(res.StatusCode, http.StatusCreated) // register "i_am_buzz"

		payload = `
		{
			"username":"i_am_burp",
			"email":"burpmail.com",
			"password":"p4$$w4rD"
		}`

		res, _ = srv.Client().Post(srv.URL+"/api/v1/account/", applicationJson, strings.NewReader(payload))
		is.Equal(res.StatusCode, http.StatusBadRequest) // registration failed

		res, _ = srv.Client().Get(srv.URL + "/api/v1/account/")
		is.Equal(res.StatusCode, http.StatusOK) // list some accounts

		type body struct {
			Length int `json:"length"`
		}

		var bd body
		_ = json.NewDecoder(res.Body).Decode(&bd)
		res.Body.Close()
		is.Equal(bd.Length, 2) // get the two registered accounts
	})

	t.Run("get a user by a key", func(t *testing.T) {
		sid := lastSplitValue(fizzUrl.String(), "/")
		res, _ := srv.Client().Get(srv.URL + "/api/v1/account/" + sid)
		is.Equal(res.StatusCode, http.StatusOK) // get a user by suid
	})

	type token struct {
		IDToken     string `json:"idToken"`
		AccessToken string `json:"accessToken"`
	}

	var fizzTk token
	fizzC := &http.Cookie{}
	t.Run(`sign-in with an account`, func(t *testing.T) {
		payload := `
		{
			"email":"buzz@mail.com",
			"password":"fizz_$PW_10"
		}`

		res, _ := srv.Client().Post(srv.URL+"/api/v1/auth/", applicationJson, strings.NewReader(payload))
		is.Equal(res.StatusCode, http.StatusNotFound) // invalid email

		payload = `
		{
			"email":"fizz@mail.com",
			"password":"fizz_$PW_10"
		}`

		res, _ = srv.Client().Post(srv.URL+"/api/v1/auth/", applicationJson, strings.NewReader(payload))
		is.Equal(res.StatusCode, http.StatusNotFound) // invalid password

		payload = `
		{
			"email":"fizz@mail.com",
			"password":"p4$$w4rD"
		}`

		res, _ = srv.Client().Post(srv.URL+"/api/v1/auth/", applicationJson, strings.NewReader(payload))
		is.Equal(res.StatusCode, http.StatusOK) // successful sign-in

		err := json.NewDecoder(res.Body).Decode(&fizzTk)
		res.Body.Close()
		is.NoErr(err) // parsing json with tokens

		for _, k := range res.Cookies() {
			t.Log(k.Value)
			if k.Name == cookieName {
				fizzC = k
			}
		}
	})

	t.Run(`access authorized endpoints`, func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/account/me", nil)
		req.Header.Set(`Authorization`, fmt.Sprintf(`Bearer %s`, fizzTk.AccessToken))
		res, _ := srv.Client().Do(req)
		is.Equal(res.StatusCode, http.StatusOK) // authorized endpoint
	})

	t.Run(`refresh token for "i_am_fizz"`, func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/auth/", nil)
		req.AddCookie(fizzC)

		res, _ := srv.Client().Do(req)
		is.Equal(res.StatusCode, http.StatusOK) // refresh token
	})
}

func lastSplitValue(s, substr string) string {
	return s[strings.LastIndex(s, substr)+1:]
}
