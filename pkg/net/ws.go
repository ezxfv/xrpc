package net

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/edenzhong7/xrpc/pkg/log"

	ws "github.com/gorilla/websocket"
)

const (
	wsPath = "/xrpc"
)

var (
	wsDialer Dialer = func(ctx context.Context, addr string) (conn Conn, err error) {
		hp := strings.Split(addr, ":")
		wsServer := genWsURL(hp[0], hp[1])
		c, _, err := ws.DefaultDialer.Dial(wsServer, nil)
		if err != nil {
			log.Fatal("ws dial:", err)
		}
		conn = newWSConn(c)
		return
	}

	_ Listener = &WSListener{}
	_ Conn     = &WSConnection{}
	// Default gorilla upgrader
	upgrader = ws.Upgrader{
		// Allow requests from *all* origins.
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	connID int64 = 0
)

func init() {
	RegisterDialer(WS, wsDialer)
	RegisterListenerBuilder(WS, newWSListener)
}

func genWsURL(host, port string) string {
	u := url.URL{Scheme: "ws", Host: host + ":" + port, Path: wsPath}
	return u.String()
}

func newWSListener(ctx context.Context, addr string) (lis Listener, err error) {
	wsListener := &WSListener{
		addr:     &XAddr{network: WS, addr: addr},
		connChan: make(chan Conn, 1024),
		closed:   false,
	}
	go wsListener.listen()
	lis = wsListener
	return
}

func newConnID() int64 {
	atomic.AddInt64(&connID, 1)
	return connID
}

func newWSConn(c *ws.Conn) (conn Conn) {
	wsConn := &WSConnection{
		id:   newConnID(),
		Conn: c,
		mu:   &sync.Mutex{},
	}
	conn = wsConn
	return
}

type WSListener struct {
	addr     Addr
	connChan chan Conn
	closed   bool
	server   *http.Server
}

func (wsl *WSListener) Accept() (Conn, error) {
	return wsl.AcceptFullConn()
}

func (wsl *WSListener) Addr() net.Addr {
	return wsl.addr
}

func (wsl *WSListener) handleConnection(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Debug("upgrader:", err)
		return
	}
	conn := newWSConn(c)
	select {
	case wsl.connChan <- conn:
	default:
		conn.Close()
		log.Debugf("ws %s->%s: conn chan is full", conn.RemoteAddr().String(), conn.LocalAddr().String())
	}
}

func (wsl *WSListener) listen() (err error) {
	http.HandleFunc(wsPath, wsl.handleConnection)
	server := &http.Server{
		Addr:    wsl.addr.String(),
		Handler: http.DefaultServeMux,
	}
	lis, err := TCPListen("tcp", wsl.addr.String())
	if err != nil {
		return err
	}
	wsl.server = server
	go server.Serve(lis)
	return
}

func (wsl *WSListener) AcceptFullConn() (conn Conn, err error) {
	var ok bool
	conn, ok = <-wsl.connChan
	if !ok {
		return nil, errors.New("unexpect failure")
	}
	return
}

func (wsl *WSListener) AcceptWithTimeout(timeout time.Duration) (conn Conn, err error) {
	timer := time.NewTimer(timeout)
	var ok bool
	select {
	case conn, ok = <-wsl.connChan:
		if !ok {
			return nil, errors.New("get conn from ws conn chan failed")
		}
		return
	case <-timer.C:
		return nil, errors.New("accept WS Connect timeout")
	}
}

func (wsl *WSListener) Close() (err error) {
	wsl.closed = true
	close(wsl.connChan)
	for v := range wsl.connChan {
		err = v.(Conn).Close()
		if err != nil {
			return
		}
	}
	if wsl.server != nil {
		err = wsl.server.Close()
	}
	return err
}

type WSConnection struct {
	*ws.Conn
	id       int64
	readBuf  []byte
	writeBuf []byte
	closed   bool
	mu       *sync.Mutex
}

func (wsc *WSConnection) ID() int64 {
	return wsc.id
}

func (wsc *WSConnection) Read(b []byte) (n int, err error) {
	_, message, err := wsc.Conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	return copy(b, message), nil
}

func (wsc *WSConnection) Write(b []byte) (n int, err error) {
	err = wsc.Conn.WriteMessage(ws.BinaryMessage, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (wsc *WSConnection) SetDeadline(t time.Time) (err error) {
	err = wsc.Conn.SetWriteDeadline(t)
	err = wsc.Conn.SetReadDeadline(t)
	return
}

func (wsc *WSConnection) Close() error {
	wsc.readBuf = nil
	wsc.writeBuf = nil
	err := wsc.Conn.WriteMessage(ws.CloseMessage, ws.FormatCloseMessage(ws.CloseNormalClosure, ""))
	if err != nil {
		return err
	}
	return wsc.Conn.Close()
}

func (wsc *WSConnection) SupportMux() bool {
	return false
}
