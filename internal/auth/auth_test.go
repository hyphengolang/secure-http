package auth

import (
	"testing"
	"time"

	"github.com/hyphengolang/prelude/testing/is"
	"github.com/hyphengolang/prelude/types/password"
	"github.com/hyphengolang/prelude/types/suid"

	"secure.adoublef.com/internal"
)

func TestToken(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	t.Run(`generate & sign tokens`, func(t *testing.T) {
		private, public := RS256()
		t.Log(private)
		t.Log(public)

		u := internal.User{
			ID:       suid.NewUUID(),
			Username: "fizz_user",
			Email:    "fizz@mail.com",
			Password: password.Password("492045rf-vf").MustHash(),
		}

		o := SignOption{
			Issuer:     "api.adoublef.com",
			Subject:    u.ID.ShortUUID().String(),
			Audience:   []string{"http://www.adoublef.com", "https://www.adoublef.com"},
			Claims:     map[string]any{"email": u.Email},
			Expiration: time.Hour * 10,
		}

		signed, err := Sign(private, &o)
		is.NoErr(err) // sign id token

		_, err = Parse(public, signed)
		is.NoErr(err) // parse token
	})
}
