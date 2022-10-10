package service

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"secure.adoublef.com/service/user"
	"secure.adoublef.com/store"
)

func (s Service) routes() {
	s.m.Use(middleware.Logger)
}

func (s Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

type Service struct {
	m chi.Router
}

func New(ctx context.Context, st *store.Store) http.Handler {
	s := &Service{m: chi.NewMux()}
	s.routes()

	user.NewService(ctx, s.m, st.UserRepo())
	return s
}
