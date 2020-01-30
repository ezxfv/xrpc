package net

import (
	"context"

	_ "github.com/lesismal/nbio"
)

func init() {
	RegisterDialer(NB, nbDialer)
	RegisterListenerBuilder(NB, newNbListener)
}

var (
	nbDialer Dialer = func(ctx context.Context, addr string) (conn Conn, err error) {
		return
	}
)

func newNbListener(ctx context.Context, addr string) (lis Listener, err error) {
	return
}

type nbConn struct {
	Conn
}

type nbListener struct {
}
