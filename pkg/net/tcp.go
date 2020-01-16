package net

import (
	"context"
	"net"
)

var (
	tcpDialer Dialer = func(ctx context.Context, addr string) (conn Conn, err error) {
		return net.Dial("tcp", addr)
	}
)

func init() {
	RegisterDialer(TCP, tcpDialer)
	RegisterListenerBuilder(TCP, newTCPListener)
}

func newTCPListener(ctx context.Context, addr string) (lis Listener, err error) {
	lis, err = net.Listen("tcp", addr)
	return
}
