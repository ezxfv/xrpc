package net

import (
	"net"

	reuseport "github.com/libp2p/go-reuseport"
)

type (
	PacketConn   = net.PacketConn
	TCPListener  = net.TCPListener
	UnixListener = net.UnixListener
)

var (
	TCPListen       = reuseport.Listen
	UDPListen       = reuseport.ListenPacket
	ResolveTCPAddr  = net.ResolveTCPAddr
	ResolveUDPAddr  = net.ResolveUDPAddr
	ResolveUnixAddr = net.ResolveUnixAddr
)
