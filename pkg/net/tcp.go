package net

var (
	tcpDialer Dialer = func(addr string) (conn Conn, err error) {
		return
	}
)

func init() {
	RegisterDialer(TCP, tcpDialer)
}
