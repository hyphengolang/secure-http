package websocket

import (
	"context"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Conn interface {
	Close() error

	WriteString(p string) error
	ReadString() (string, error)
}

type serverConn struct {
	rwc net.Conn
}

func (c serverConn) Close() error { return c.rwc.Close() }

func (c serverConn) ReadString() (string, error) {
	b, err := wsutil.ReadClientText(c.rwc)
	return string(b), err
}

func (c serverConn) WriteString(s string) error {
	return wsutil.WriteServerText(c.rwc, []byte(s))
}

type clientConn struct {
	rwc net.Conn
}

func (c clientConn) Close() error { return c.rwc.Close() }

func (c clientConn) ReadString() (string, error) {
	b, err := wsutil.ReadServerText(c.rwc)
	return string(b), err
}

func (c clientConn) WriteString(s string) error {
	return wsutil.WriteClientText(c.rwc, []byte(s))
}

func UpgradeHTTP(w http.ResponseWriter, r *http.Request) (c Conn, err error) {
	rwc, _, _, err := ws.UpgradeHTTP(r, w)
	return serverConn{rwc}, err
}

func Dial(ctx context.Context, urlStr string) (c Conn, err error) {
	conn, _, _, err := ws.Dial(context.Background(), urlStr)
	return clientConn{conn}, err
}
