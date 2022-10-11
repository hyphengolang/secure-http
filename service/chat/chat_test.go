package chat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/hyphengolang/prelude/testing/is"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

const (
	applicationJson = "application/json"
)

var h http.Handler

func init() {
	h = NewService(context.Background(), chi.NewMux())
}

func TestService(t *testing.T) {
	t.Parallel()
	is, ctx := is.New(t), context.TODO()
	_, _ = is, ctx

	srv := httptest.NewServer(h)
	t.Cleanup(func() { srv.Close() })

	t.Run("echo from server", func(t *testing.T) {
		payload := `Hello`

		conn, _, _, err := ws.Dial(context.Background(), "ws"+strings.TrimPrefix(srv.URL, "http")+"/api/v1/chat/")
		is.NoErr(err) // failed to upgrade
		t.Cleanup(func() { conn.Close() })

		err = wsutil.WriteClientText(conn, []byte(payload))
		is.NoErr(err) // write to server

		b, _, err := wsutil.ReadServerData(conn)
		is.NoErr(err) // reading echo

		is.Equal(string(b), payload)
		is.True(string(b) != `World`)
	})
}
