package chat

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	www "github.com/hyphengolang/prelude/http"
	"github.com/hyphengolang/prelude/http/websocket"
)

type Map struct {
	sync.RWMutex
	cs  map[websocket.Conn]struct{}
	seq uint
	cap uint
}

func (m *Map) Size() uint { return m.seq }

func NewMap(cap uint) Map {
	return Map{
		cap: cap,
		cs:  make(map[websocket.Conn]struct{}),
	}
}

func (m *Map) Range(f func(conn websocket.Conn) error) error {
	m.RLock()
	defer m.RUnlock()
	for c := range m.cs {
		if err := f(c); err != nil {
			return err
		}
	}

	return nil
}

func (m *Map) BroadCast(p []byte) error {
	m.RLock()
	for c := range m.cs {
		if err := c.Write(p); err != nil {
			return err
		}
	}

	m.RUnlock()
	return nil
}

func (m *Map) Delete(c websocket.Conn) error {
	m.Lock()
	delete(m.cs, c)
	m.Unlock()
	m.seq--
	return c.Close()
}

func (m *Map) Store(c websocket.Conn) {
	m.Lock()
	m.cs[c] = struct{}{}
	m.Unlock()
	m.seq++
}

/*
https://github.dev/gorilla/websocket/blob/master/examples/echo/server.go
*/
func (s Service) routes() {
	m := NewMap(4)

	s.m.Route("/api/v1/chat", func(r chi.Router) {
		r.Get("/", s.handleChat(&m))
	})
}

func (s Service) handleChat(m *Map) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if m.Size() >= 2 {
			s.respond(w, r, errors.New("full capacity"), http.StatusBadRequest)
			return
		}

		conn, err := websocket.UpgradeHTTP(w, r)
		if err != nil {
			s.respond(w, r, err, http.StatusUpgradeRequired)
			return
		}

		m.Store(conn)
		defer m.Delete(conn)

		for {
			msg, _ := conn.Read()
			_ = m.Range(func(conn websocket.Conn) error { return conn.Write(msg) })
		}
	}
}

type Service struct {
	m chi.Router

	respond   func(w http.ResponseWriter, r *http.Request, data any, status int)
	decode    func(w http.ResponseWriter, r *http.Request, data any) (err error)
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
