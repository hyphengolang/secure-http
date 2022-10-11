package chat

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	www "github.com/hyphengolang/prelude/http"
)

func (s Service) routes() {
	u := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	s.m.Route("/api/v1/chat", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			conn, err := u.Upgrade(w, r, nil)
			if err != nil {
				s.respond(w, r, err, http.StatusUpgradeRequired)
				return
			}

			var s string
			_ = conn.ReadJSON(&s)
			_ = conn.WriteJSON("Hello " + s + "!")
		})
	})
}

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

func (s Service) respondText(w http.ResponseWriter, r *http.Request, status int) {
	s.respond(w, r, http.StatusText(status), status)
}
