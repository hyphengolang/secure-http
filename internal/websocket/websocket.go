package websocket

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Conn interface {
	Close() error
	State() ws.State

	Read() ([]byte, error)
	ReadString() (string, error)
	ReadJSON(v any) error

	Write(p []byte) error
	WriteString(p string) error
	WriteJSON(v any) error
}

type serverConn struct {
	r *wsutil.Reader
	w *wsutil.Writer

	rwc net.Conn
}

func (c serverConn) State() ws.State { return ws.StateServerSide }

func (c serverConn) Close() error { return c.rwc.Close() }

func (c serverConn) Read() ([]byte, error) {
	b, err := wsutil.ReadClientBinary(c.rwc)
	return b, err
}

func (c serverConn) ReadString() (string, error) {
	h, err := c.r.NextFrame()
	if err != nil {
		return "", err
	}

	// Reset writer to write frame with right operation code.
	c.w.Reset(c.rwc, c.State(), h.OpCode)

	b, err := io.ReadAll(c.r)
	return string(b), err
}

func (c serverConn) ReadJSON(v any) error {
	// NOTE read next frame goes in here for now
	h, err := c.r.NextFrame()
	if err != nil {
		return err
	}

	if h.OpCode == ws.OpClose {
		return io.EOF
	}
	return json.NewDecoder(c.r).Decode(v)
}

func (c serverConn) Write(p []byte) error {
	return wsutil.WriteMessage(c.rwc, c.State(), ws.OpBinary, p)
}

func (c serverConn) WriteString(s string) error {
	_, err := io.WriteString(c.w, s)
	if err != nil {
		return err
	}
	return c.w.Flush()
}

func (c serverConn) WriteJSON(v any) error {
	if err := json.NewEncoder(c.w).Encode(v); err != nil {
		return err
	}
	return c.w.Flush()
}

type clientConn struct {
	rwc net.Conn
}

func (c clientConn) State() ws.State { return ws.StateClientSide }

func (c clientConn) Close() error { return c.rwc.Close() }

func (c clientConn) Read() ([]byte, error) {
	b, err := wsutil.ReadServerText(c.rwc)
	return b, err
}

func (c clientConn) ReadString() (string, error) {
	p, _, err := wsutil.ReadData(c.rwc, c.State())
	// b, err := wsutil.ReadServerText(c.rwc)
	return string(p), err
}

func (c clientConn) ReadJSON(v any) error {
	r := wsutil.NewReader(c.rwc, c.State())

	// NOTE read next frame goes in here for now
	h, err := r.NextFrame()
	if err != nil {
		return err
	}
	if h.OpCode == ws.OpClose {
		return io.EOF
	}
	return json.NewDecoder(r).Decode(v)
}

func (c clientConn) Write(p []byte) error {
	return wsutil.WriteMessage(c.rwc, c.State(), ws.OpBinary, p)
}

func (c clientConn) WriteString(s string) error {
	return wsutil.WriteMessage(c.rwc, c.State(), ws.OpText, []byte(s))
}

func (c clientConn) WriteJSON(v any) error {
	w := wsutil.NewWriter(c.rwc, c.State(), ws.OpText)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		return err
	}
	return w.Flush()
}

func UpgradeHTTP(w http.ResponseWriter, r *http.Request) (c Conn, err error) {
	rwc, _, _, err := ws.UpgradeHTTP(r, w)
	// TODO remove the need to define reader/write on the connection
	return serverConn{
		r:   wsutil.NewReader(rwc, ws.StateServerSide),
		w:   wsutil.NewWriter(rwc, ws.StateServerSide, ws.OpText),
		rwc: rwc,
	}, err
}

func Dial(ctx context.Context, urlStr string) (c Conn, err error) {
	rwc, _, _, err := ws.Dial(context.Background(), urlStr)
	// TODO remove the need to define reader/write on the connection
	return clientConn{rwc}, err
}
