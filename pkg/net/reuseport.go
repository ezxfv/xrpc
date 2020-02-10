package net

import (
	reuseport "github.com/libp2p/go-reuseport"
)

var (
	TCPListen = reuseport.Listen
	UDPListen = reuseport.ListenPacket
)
