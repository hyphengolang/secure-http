package chat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/hyphengolang/prelude/http/websocket"
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

func newConn(urlStr string) (websocket.Conn, error) {
	return websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(urlStr, "http"))
}

func TestService(t *testing.T) {
	t.Parallel()
	is, ctx := is.New(t), context.TODO()
	_, _ = is, ctx

	srv := httptest.NewServer(h)
	t.Cleanup(func() { srv.Close() })

	t.Run("send a json message as bytes", func(t *testing.T) {
		c1, err := newConn(srv.URL + "/api/v1/chat")
		is.NoErr(err) // c1 connected

		c2, err := newConn(srv.URL + "/api/v1/chat")
		is.NoErr(err) // c2 connected

		_, err = newConn(srv.URL + "/api/v1/chat")
		is.True(err != nil) // full capacity

		t.Cleanup(func() { c1.Close(); c2.Close() })

		type payload struct {
			MsgTyp string
			Data   string
		}

		input := `
{ 
	"type":"echo",
	"data":"World" 
}`

		err = c1.Write([]byte(input))
		is.NoErr(err) // write to server

		// var output string
		output, err := c2.Read()
		is.NoErr(err) // reading echo

		var pl payload
		err = json.Unmarshal(output, &pl)
		is.NoErr(err) // decoding message on client

		is.Equal(pl.Data, `World`) // c2 receives message from c1
	})
}
