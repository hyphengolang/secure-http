package chat

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	www "github.com/hyphengolang/prelude/http"
)

/*
https://github.dev/gorilla/websocket/blob/master/examples/echo/server.go
*/
func (s Service) routes() {
	s.m.Route("/api/v1/chat", func(r chi.Router) {
		r.Get("/", s.handleChat())
	})
}

func (s Service) handleChat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rwc, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			s.respond(w, r, err, http.StatusUpgradeRequired)
			return
		}

		defer rwc.Close()

		for {
			msg, _ := wsutil.ReadClientText(rwc)
			_ = wsutil.WriteServerText(rwc, msg)
		}
	}
}

/* WEBSOCKET */

type Service struct {
	m chi.Router

	respond   func(w http.ResponseWriter, r *http.Request, data any, status int)
	decode    func(rw http.ResponseWriter, r *http.Request, data any) (err error)
	created   func(w http.ResponseWriter, r *http.Request, id string)
	setCookie func(w http.ResponseWriter, cookie *http.Cookie)

	log  func(v ...any)
	logf func(format string, v ...any)
}

func NewService(ctx context.Context, m chi.Router) http.Handler {
	s := Service{
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

func (s Service) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.m.ServeHTTP(w, r) }

func (s Service) respondStatus(w http.ResponseWriter, r *http.Request, status int) {
	s.respond(w, r, http.StatusText(status), status)
}
