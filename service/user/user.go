package user

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	www "github.com/hyphengolang/prelude/http"
	"github.com/lestrrat-go/jwx/v2/jwk"

	"secure.adoublef.com/internal"
	"secure.adoublef.com/internal/auth"
	"secure.adoublef.com/internal/suid"
)

/*
Register to service

	[?] POST /api/v1/account/

Get list of accounts

	[?] GET /api/v1/account/

Get current user info

	[ ] GET /api/v1/account/me

Delete my account

	[ ] DELETE /api/v1/account/me

Get a user's info by uuid

	[ ] GET /api/v1/account/{uuid}

Sign in with credentials

	[?] POST /api/v1/token

Sign out with credentials

	[ ] POST /api/v1/token

Refresh token

	[ ] GET /api/v1/token
*/
func (s Service) routes() {
	private, public := auth.RS256()

	s.m.Route("/api/v1/account", func(r chi.Router) {
		r.Post("/", s.handleCreateAccount())

		// authorization required
		r.Get("/", s.handleGetAccountList())
		r.Get("/{uuid}", s.handleGetAccount())
		r.Get("/me", s.handleGetMyAccount(public))
		r.Delete("/me", s.handleSignOut())
	})

	s.m.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/", s.handleSignIn(private))

		// authorization required
		r.Delete("/", http.NotFound)
		r.Get("/", s.handleRefreshToken(private, public))
	})
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
		s.respondText(w, r, http.StatusOK)
	}
}

func (s Service) handleRefreshToken(private, public jwk.Key) http.HandlerFunc {
	type token struct {
		AccessToken string `json:"accessToken"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// auth middleware
		jtk, err := auth.ParseCookie(r, public, cookieName)
		if err != nil {
			s.respond(w, r, err, http.StatusUnauthorized)
			return
		}

		email, ok := jtk.PrivateClaims()["email"].(string)
		if !ok {
			s.respondText(w, r, http.StatusInternalServerError)
			return
		}
		// auth middleware
		u, err := s.r.Select(r.Context(), internal.Email(email))
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
		var d User
		if err := s.decode(w, r, &d); err != nil {
			s.respond(w, r, err, http.StatusBadRequest)
			return
		}

		u, err := s.r.Select(r.Context(), d.Email)
		if err != nil {
			s.respond(w, r, err, http.StatusNotFound)
			return
		}

		if err := u.Password.Compare(d.Password.String()); err != nil {
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

func (s Service) handleGetAccount() http.HandlerFunc {
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

func (s Service) handleGetMyAccount(public jwk.Key) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// auth middleware
		tk, err := auth.ParseRequest(r, public)
		if err != nil {
			s.respond(w, r, err, http.StatusUnauthorized)
			return
		}

		email, ok := tk.PrivateClaims()["email"].(string)
		if !ok {
			s.respondText(w, r, http.StatusInternalServerError)
			return
		}
		// auth middleware

		me, err := s.r.Select(r.Context(), internal.Email(email))
		if err != nil {
			s.respond(w, r, err, http.StatusNotFound)
			return
		}

		s.respond(w, r, me, http.StatusOK)
	}
}

func (s Service) handleGetAccountList() http.HandlerFunc {
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

func (s Service) handleCreateAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u internal.User
		if err := s.newUser(w, r, &u); err != nil {
			s.respond(w, r, err, http.StatusBadRequest)
			return
		}

		if err := s.r.Insert(r.Context(), &u); err != nil {
			s.respond(w, r, err, http.StatusInternalServerError)
			return
		}

		s.created(w, r, u.ID.ShortUUID().String())
	}
}

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

func (s Service) newUser(w http.ResponseWriter, r *http.Request, u *internal.User) error {
	var d User
	if err := s.decode(w, r, &d); err != nil {
		return err
	}

	h, err := d.Password.Hash()
	if err != nil {
		return err
	}

	*u = internal.User{
		ID:       suid.NewUUID(),
		Username: d.Username,
		Email:    d.Email,
		Password: h,
	}

	return nil
}

type User struct {
	ID       suid.UUID         `json:"id"`
	Username string            `json:"username"`
	Email    internal.Email    `json:"email"`
	Password internal.Password `json:"password"`
}

type Service struct {
	ctx context.Context

	m         chi.Router
	respond   func(w http.ResponseWriter, r *http.Request, data any, status int)
	decode    func(rw http.ResponseWriter, r *http.Request, data any) (err error)
	created   func(w http.ResponseWriter, r *http.Request, id string)
	setCookie func(w http.ResponseWriter, cookie *http.Cookie)

	r internal.UserRepo

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
		m:         m,
		respond:   www.Respond,
		decode:    www.Decode,
		created:   www.Created,
		setCookie: http.SetCookie,
		r:         r,
		log:       log.Println,
		logf:      log.Printf,
	}
	s.routes()
	return s
}

const (
	cookieName = "__adf"
)
