package net

import (
	"context"
	"net"
)

var (
	// addr example: /tmp/xrpc.sock
	unixDialer Dialer = func(ctx context.Context, addr string) (conn Conn, err error) {
		tc, err := net.Dial("unix", addr)
		conn = &unixConn{tc}
		return
	}
)

func init() {
	RegisterDialer(UNIX, unixDialer)
	RegisterListenerBuilder(UNIX, newUnixListener)
}

func newUnixListener(ctx context.Context, addr string) (lis Listener, err error) {
	uaddr, err := net.ResolveUnixAddr("unix", addr)
	if err != nil {
		return nil, err
	}
	l, err := net.ListenUnix("unix", uaddr)
	lis = &unixListener{l}
	return
}

type unixConn struct {
	net.Conn
}

func (tc *unixConn) SupportMux() bool {
	return true
}

type unixListener struct {
	lis *net.UnixListener
}

func (tl *unixListener) Accept() (conn Conn, err error) {
	c, err := tl.lis.AcceptUnix()
	if err != nil {
		return
	}
	conn = &unixConn{c}
	return
}

func (tl *unixListener) Close() error {
	return tl.lis.Close()
}

func (tl *unixListener) Addr() Addr {
	return tl.lis.Addr()
}
