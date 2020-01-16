package net

import (
	"context"
	"net"
)

var (
	tcpDialer Dialer = func(ctx context.Context, addr string) (conn Conn, err error) {
		tc, err := net.Dial("tcp", addr)
		conn = &tcpConn{tc}
		return
	}
)

func init() {
	RegisterDialer(TCP, tcpDialer)
	RegisterListenerBuilder(TCP, newTCPListener)
}

func newTCPListener(ctx context.Context, addr string) (lis Listener, err error) {
	l, err := net.Listen("tcp", addr)
	lis = &tcpListener{l}
	return
}

type tcpConn struct {
	net.Conn
}

func (tc *tcpConn) SupportMux() bool {
	return true
}

type tcpListener struct {
	lis net.Listener
}

func (tl *tcpListener) Accept() (conn Conn, err error) {
	c, err := tl.lis.Accept()
	conn = &tcpConn{c}
	return
}

func (tl *tcpListener) Close() error {
	return tl.lis.Close()
}

func (tl *tcpListener) Addr() Addr {
	return tl.lis.Addr()
}
