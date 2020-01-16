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

	"github.com/edenzhong7/xrpc/pkg/algs"
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
			log.GLogger().Fatal("ws dial:", err)
		}
		conn = newWSConn(c)
		//n, err := conn.Write([]byte(wsPath))
		//if err != nil || n != len(wsPath) {
		//	return nil, errors.New("write ws path failed")
		//}
		return
	}

	_ Listener = &WSListener{}
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
		addr:     &XAddr{protocol: WS, addr: addr},
		listener: algs.NewQueue(),
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
	listener *algs.Queue
	closed   bool
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
		log.GLogger().Debug("upgrader:", err)
		return
	}
	wsl.listener.Append(c)
}

func (wsl *WSListener) listen() (err error) {
	http.HandleFunc(wsPath, wsl.handleConnection)
	return http.ListenAndServe(wsl.addr.String(), nil)
}

func (wsl *WSListener) AcceptFullConn() (conn Conn, err error) {
	wsConn, ok := wsl.listener.Pop().(*ws.Conn)
	if !ok {
		return nil, errors.New("unexpect failure")
	}
	conn = newWSConn(wsConn)
	//b := make([]byte, len(wsPath))
	//n, err := conn.Read(b)
	//if err != nil || n != len(wsPath) {
	//	return nil, errors.New("read ws path failed")
	//}
	return
}

func (wsl *WSListener) AcceptWithTimeout(timeout time.Duration) (conn Conn, err error) {
	timer := time.NewTimer(timeout)
	c := make(chan interface{})
	go func() {
		conn, err = wsl.AcceptFullConn()
		c <- nil
	}()
	select {
	case <-c:
		return
	case <-timer.C:
		return nil, errors.New("accept WS Connect timeout")
	}
}

func (wsl *WSListener) Close() (err error) {
	wsl.closed = true
	for _, v := range wsl.listener.Items() {
		err = v.(*ws.Conn).Close()
		if err != nil {
			return
		}
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
