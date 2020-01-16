package net

import "net"

var (
	tcpDialer Dialer = func(addr string) (conn Conn, err error) {
		return net.Dial("tcp", addr)
	}
)

func init() {
	RegisterDialer(TCP, tcpDialer)
}
