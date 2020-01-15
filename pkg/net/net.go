package net

import "net"

type Protocol = string

const (
	TCP  Protocol = "tcp"
	UDP           = "udp"
	QUIC          = "quic"
	WS            = "ws"
	KCP           = "kcp"
)

type Listener interface {
	net.Listener
}

type Conn interface {
	net.Conn
}

func Listen(protocol Protocol, addr string) (lis Listener, err error) {
	return
}

type Dialer struct {
}
