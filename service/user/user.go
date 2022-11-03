package user

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	www "github.com/hyphengolang/prelude/http"
	"github.com/hyphengolang/prelude/types/email"
	"github.com/hyphengolang/prelude/types/password"
	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/hyphengolang/prelude/types/suid"
	"secure.adoublef.com/internal"
	"secure.adoublef.com/internal/auth"
)

/*
Register to service

	[?] POST /api/v1/user/

Get list of accounts

	[?] GET /api/v1/user/

Get current user info

	[ ] GET /api/v1/user/me

Delete my account

	[ ] DELETE /api/v1/user/me

Get a user's info by uuid

	[ ] GET /api/v1/user/{uuid}

Sign in with credentials

	[?] POST /api/v1/token

Sign out with credentials

	[ ] POST /api/v1/token

Refresh token

	[ ] GET /api/v1/token
*/
func (s Service) routes() {
	private, public := auth.RS256()

	s.m.Route("/api/v1/user", func(r chi.Router) {
		r.Post("/", s.handleRegisterUser())

		r.Get("/", s.handleListRegisteredUsers())
		r.Get("/{uuid}", s.handleSearchUser())
		r.Get("/me", s.handleGetMyDetails(public))

		r.Delete("/me", s.handleDeregisterUser(public))
	})

	s.m.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/", s.handleSignIn(private))
		r.Delete("/", s.handleSignOut())
		r.Get("/", s.handleRefreshToken(private, public))
	})
}

func (s Service) handleDeregisterUser(public jwk.Key) http.HandlerFunc {
	// auth required
	return func(w http.ResponseWriter, r *http.Request) {
		tk, err := auth.ParseRequest(r, public)
		if err != nil {
			s.respond(w, r, err, http.StatusUnauthorized)
			return
		}

		e, ok := tk.PrivateClaims()["email"].(string)
		if !ok {
			s.respondText(w, r, http.StatusInternalServerError)
			return
		}

		me, err := s.r.Select(r.Context(), email.Email(e))
		if err != nil {
			s.respond(w, r, err, http.StatusNotFound)
			return
		}

		if err := s.r.Delete(r.Context(), me.ID); err != nil {
			s.respond(w, r, err, http.StatusInternalServerError)
			return
		}

		c := &http.Cookie{
			Path:     "/",
			Name:     cookieName,
			HttpOnly: true,
			MaxAge:   -1,
		}

		s.setCookie(w, c)
		s.respond(w, r, nil, http.StatusNoContent)
	}
}

func (s Service) handleSignOut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := &http.Cookie{
			Path:     "/",
			Name:     cookieName,
			HttpOnly: true,
			MaxAge:   -1,
		}

		s.setCookie(w, c)
		s.respondText(w, r, http.StatusNoContent)
	}
}

func (s Service) handleRefreshToken(private, public jwk.Key) http.HandlerFunc {
	type token struct {
		AccessToken string `json:"accessToken"`
	}

	// auth required
	return func(w http.ResponseWriter, r *http.Request) {
		jtk, err := auth.ParseCookie(r, public, cookieName)
		if err != nil {
			s.respond(w, r, err, http.StatusUnauthorized)
			return
		}

		e, ok := jtk.PrivateClaims()["email"].(string)
		if !ok {
			s.respondText(w, r, http.StatusInternalServerError)
			return
		}

		u, err := s.r.Select(r.Context(), email.Email(e))
		if err != nil {
			s.respond(w, r, err, http.StatusForbidden)
			return
		}

		_, ats, _, err := s.signedTokens(private, u)
		if err != nil {
			s.respond(w, r, err, http.StatusInternalServerError)
			return
		}

		tk := token{
			AccessToken: string(ats),
		}

		s.respond(w, r, tk, http.StatusOK)
	}
}

func (s Service) handleSignIn(private jwk.Key) http.HandlerFunc {
	type payload struct {
		IDToken     string `json:"idToken"`
		AccessToken string `json:"accessToken"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var dto User
		if err := s.decode(w, r, &dto); err != nil {
			s.respond(w, r, err, http.StatusBadRequest)
			return
		}

		u, err := s.r.Select(r.Context(), dto.Email)
		if err != nil {
			s.respond(w, r, err, http.StatusNotFound)
			return
		}

		if err := u.Password.Compare(dto.Password.String()); err != nil {
			s.respond(w, r, err, http.StatusForbidden)
			return
		}

		its, ats, rts, err := s.signedTokens(private, u)
		if err != nil {
			s.respond(w, r, err, http.StatusInternalServerError)
			return
		}

		rc := &http.Cookie{
			Path:     "/",
			Name:     cookieName,
			Value:    string(rts),
			HttpOnly: true,
			Secure:   r.TLS != nil,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   30 * 24 * 7,
		}

		s.setCookie(w, rc)

		p := &payload{
			IDToken:     string(its),
			AccessToken: string(ats),
		}

		s.respond(w, r, p, http.StatusOK)
	}
}

func (s Service) handleSearchUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := s.parseUUID(w, r)
		if err != nil {
			s.respond(w, r, err, http.StatusBadRequest)
			return
		}

		u, err := s.r.Select(r.Context(), uid)
		if err != nil {
			s.respond(w, r, err, http.StatusNotFound)
			return
		}

		s.respond(w, r, u, http.StatusOK)
	}
}

func (s Service) handleGetMyDetails(public jwk.Key) http.HandlerFunc {
	// auth required
	return func(w http.ResponseWriter, r *http.Request) {
		tk, err := auth.ParseRequest(r, public)
		if err != nil {
			s.respond(w, r, err, http.StatusUnauthorized)
			return
		}

		e, ok := tk.PrivateClaims()["email"].(string)
		if !ok {
			s.respondText(w, r, http.StatusInternalServerError)
			return
		}

		me, err := s.r.Select(r.Context(), email.Email(e))
		if err != nil {
			s.respond(w, r, err, http.StatusNotFound)
			return
		}

		s.respond(w, r, me, http.StatusOK)
	}
}

func (s Service) handleListRegisteredUsers() http.HandlerFunc {
	type payload struct {
		Length int             `json:"length"`
		Data   []internal.User `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		us, err := s.r.SelectMany(r.Context())
		if err != nil {
			s.respond(w, r, err, http.StatusInternalServerError)
			return
		}

		p := payload{
			Length: len(us),
			Data:   us,
		}

		s.respond(w, r, p, http.StatusOK)
	}
}

func (s Service) handleRegisterUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := s.newUser(w, r)
		if err != nil {
			s.respond(w, r, err, http.StatusBadRequest)
			return
		}

		if err := s.r.Insert(r.Context(), u); err != nil {
			s.respond(w, r, err, http.StatusInternalServerError)
			return
		}

		s.created(w, r, u.ID.ShortUUID().String())
	}
}

// Service Utils

func (s Service) signedTokens(private jwk.Key, u *internal.User) (its, ats, rts []byte, err error) {
	o := auth.SignOption{
		Issuer:   "api.adoublef.com",
		Subject:  suid.NewUUID().String(),
		Audience: []string{"http://www.adoublef.com", "https://www.adoublef.com"},
		Claims:   map[string]any{"email": u.Email, "id": u.ID.ShortUUID(), "username": u.Username},
	}

	// its
	o.Expiration = time.Hour * 10
	if its, err = auth.Sign(private, &o); err != nil {
		return
	}

	// ats
	o.Expiration = time.Minute * 5
	if ats, err = auth.Sign(private, &o); err != nil {
		return
	}

	// rts
	o.Expiration = time.Hour * 24 * 7
	if rts, err = auth.Sign(private, &o); err != nil {
		return
	}

	return
}

func (s Service) respondText(w http.ResponseWriter, r *http.Request, status int) {
	s.respond(w, r, http.StatusText(status), status)
}

func (s Service) parseUUID(w http.ResponseWriter, r *http.Request) (suid.UUID, error) {
	return suid.ParseString(chi.URLParam(r, "uuid"))
}

func (s Service) newUser(w http.ResponseWriter, r *http.Request) (*internal.User, error) {
	var dto User
	if err := s.decode(w, r, &dto); err != nil {
		return nil, err
	}

	h, err := dto.Password.Hash()
	if err != nil {
		return nil, err
	}

	u := &internal.User{
		ID:       suid.NewUUID(),
		Username: dto.Username,
		Email:    dto.Email,
		Password: h,
	}

	return u, nil
}

type User struct {
	ID       suid.UUID         `json:"id"`
	Username string            `json:"username"`
	Email    email.Email       `json:"email"`
	Password password.Password `json:"password"`
}

type Service struct {
	ctx context.Context

	r internal.UserRepo

	m         chi.Router
	respond   func(w http.ResponseWriter, r *http.Request, data any, status int)
	decode    func(rw http.ResponseWriter, r *http.Request, data any) (err error)
	created   func(w http.ResponseWriter, r *http.Request, id string)
	setCookie func(w http.ResponseWriter, cookie *http.Cookie)

	log  func(v ...any)
	logf func(format string, v ...any)
}

func (s Service) Context() context.Context {
	if s.ctx != nil {
		return s.ctx
	}

	return context.Background()
}

func (s Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

func NewService(ctx context.Context, m chi.Router, r internal.UserRepo) http.Handler {
	s := &Service{
		ctx:       ctx,
		r:         r,
		m:         m,
		respond:   www.Respond,
		decode:    www.Decode,
		created:   www.Created,
		setCookie: http.SetCookie,
		log:       log.Println,
		logf:      log.Printf,
	}
	s.routes()
	return s
}

const (
	cookieName = "__adf"
)
