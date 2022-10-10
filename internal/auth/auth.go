package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type SignOption struct {
	IssuedAt   time.Time
	Issuer     string
	Audience   []string
	Subject    string
	Expiration time.Duration
	Claims     map[string]any
}

func Parse(key jwk.Key, token []byte) (jwt.Token, error) {
	var sep jwt.SignEncryptParseOption
	switch key := key.(type) {
	case jwk.RSAPublicKey:
		sep = jwt.WithKey(jwa.RS256, key)
	case jwk.ECDSAPublicKey:
		sep = jwt.WithKey(jwa.ES256, key)
	default:
		return nil, errors.New(`unsupported encryption`)
	}

	return jwt.Parse(token, sep)
}

/*
ParseRequest searches a http.Request object for a JWT token.

Specifying WithHeaderKey() will tell it to search under a specific
header key. Specifying WithFormKey() will tell it to search under
a specific form field.

By default, "Authorization" header will be searched.

If WithHeaderKey() is used, you must explicitly re-enable searching for "Authorization" header.

	# searches for "Authorization"
	jwt.ParseRequest(req)

	# searches for "x-my-token" ONLY.
	jwt.ParseRequest(req, jwt.WithHeaderKey("x-my-token"))

	# searches for "Authorization" AND "x-my-token"
	jwt.ParseRequest(req, jwt.WithHeaderKey("Authorization"), jwt.WithHeaderKey("x-my-token"))
*/
func ParseRequest(r *http.Request, key jwk.Key) (jwt.Token, error) {
	var sep jwt.SignEncryptParseOption
	switch key := key.(type) {
	case jwk.RSAPublicKey:
		sep = jwt.WithKey(jwa.RS256, key)
	case jwk.ECDSAPublicKey:
		sep = jwt.WithKey(jwa.ES256, key)
	default:
		return nil, errors.New(`unsupported encryption`)
	}
	return jwt.ParseRequest(r, sep)
}

func ParseCookie(r *http.Request, key jwk.Key, cookieName string) (jwt.Token, error) {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return nil, err
	}

	var sep jwt.SignEncryptParseOption
	switch key := key.(type) {
	case jwk.RSAPublicKey:
		sep = jwt.WithKey(jwa.RS256, key)
	case jwk.ECDSAPublicKey:
		sep = jwt.WithKey(jwa.ES256, key)
	default:
		return nil, errors.New(`unsupported encryption`)
	}

	return jwt.Parse([]byte(c.Value), sep)
}

func Sign(key jwk.Key, o *SignOption) ([]byte, error) {
	var sep jwt.SignEncryptParseOption
	switch key := key.(type) {
	case jwk.RSAPrivateKey:
		sep = jwt.WithKey(jwa.RS256, key)
	case jwk.ECDSAPrivateKey:
		sep = jwt.WithKey(jwa.ES256, key)
	default:
		return nil, errors.New(`unsupported encryption`)
	}

	var iat time.Time
	if o.IssuedAt.IsZero() {
		iat = time.Now().UTC()
	} else {
		iat = o.IssuedAt
	}

	t, err := jwt.NewBuilder().
		Issuer(o.Issuer).
		Audience(o.Audience).
		Subject(o.Subject).
		IssuedAt(iat).
		Expiration(iat.Add(o.Expiration)).
		Build()

	if err != nil {
		return nil, err
	}

	for k, v := range o.Claims {
		if err := t.Set(k, v); err != nil {
			return nil, err
		}
	}

	return jwt.Sign(t, sep)
}

func RS256() (private, public jwk.Key) {
	raw, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	if private, err = jwk.FromRaw(raw); err != nil {
		panic(err)
	}

	if public, err = private.PublicKey(); err != nil {
		panic(err)
	}

	return
}

func ES256() (private, public jwk.Key) {
	raw, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	if private, err = jwk.FromRaw(raw); err != nil {
		panic(err)
	}

	if public, err = private.PublicKey(); err != nil {
		panic(err)
	}

	return
}
