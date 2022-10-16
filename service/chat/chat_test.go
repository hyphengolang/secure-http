package chat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/hyphengolang/prelude/testing/is"
)

const (
	applicationJson = "application/json"
)

var h http.Handler

func init() {

	h = NewService(context.Background(), chi.NewMux())
}

// https://quii.gitbook.io/learn-go-with-tests/build-an-application/websockets

func TestService(t *testing.T) {
	t.Parallel()
	is, ctx := is.New(t), context.TODO()
	_, _ = is, ctx

	srv := httptest.NewServer(h)
	t.Cleanup(func() { srv.Close() })

	conn, _, _, err := ws.Dial(context.Background(), "ws"+strings.TrimPrefix(srv.URL, "http")+"/api/v1/chat/")
	is.NoErr(err) // failed to upgrade

	t.Cleanup(func() { conn.Close() })

	t.Run("echo server", func(t *testing.T) {
		i := `Hello Foo`

		err = wsutil.WriteClientText(conn, []byte(i))
		is.NoErr(err) // write to server

		b, err := wsutil.ReadServerText(conn)
		is.NoErr(err) // reading echo

		is.Equal(string(b), i)
	})
}
