package net

import (
	"context"
	"crypto/sha1"
	"time"

	kcp "github.com/xtaci/kcp-go"

	"golang.org/x/crypto/pbkdf2"
)

var (
	kcpDialer Dialer = func(ctx context.Context, addr string) (conn Conn, err error) {
		key := pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New)
		block, _ := kcp.NewAESBlockCrypt(key)

		// wait for server to become ready
		time.Sleep(time.Second)

		// dial to the echo server
		kc, err := kcp.DialWithOptions(addr, block, 10, 3)
		conn = &kcpConn{kc}
		return
	}
)

func init() {
	RegisterDialer(KCP, kcpDialer)
	RegisterListenerBuilder(KCP, newKcpListener)
}

func newKcpListener(ctx context.Context, addr string) (lis Listener, err error) {
	key := pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New)
	block, err := kcp.NewAESBlockCrypt(key)
	if err != nil {
		return
	}
	//l, err := kcp.ListenWithOptions(addr, block, 10, 3)
	conn, err := UDPListen("udp", addr)
	if err != nil {
		return
	}
	l, err := kcp.ServeConn(block, 10, 3, conn)
	if err != nil {
		return
	}
	lis = &kcpListener{lis: l}
	return
}

type kcpConn struct {
	*kcp.UDPSession
}

func (kc *kcpConn) SupportMux() bool {
	return true
}

type kcpListener struct {
	lis *kcp.Listener
}

func (kl *kcpListener) Accept() (conn Conn, err error) {
	c, err := kl.lis.AcceptKCP()
	conn = &kcpConn{c}
	return
}

func (kl *kcpListener) Close() error {
	return kl.lis.Close()
}

func (kl *kcpListener) Addr() Addr {
	return kl.lis.Addr()
}
