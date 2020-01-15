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

var (
	dialers = make(map[Protocol]Dialer)
)

func RegisterDialer(protocol Protocol, dialer Dialer) {
	if _, ok := dialers[protocol]; !ok {
		dialers[protocol] = dialer
	}
}

func GetDialer(protocol Protocol) Dialer {
	dialer := dialers[protocol]
	return dialer
}

type Listener interface {
	net.Listener
}

type Conn interface {
	net.Conn
}

func Listen(protocol Protocol, addr string) (lis Listener, err error) {
	return
}

func Dial(protocol Protocol, addr string) (conn Conn, err error) {
	return
}

type Dialer func(addr string) (conn Conn, err error)
