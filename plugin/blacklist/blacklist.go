package blacklist

import (
	"x.io/xrpc/pkg/net"
)

func New(blacklist map[string]bool, mask []*net.IPNet) *blacklistPlugin {
	if blacklist == nil {
		blacklist = make(map[string]bool)
	}
	if mask == nil {
		mask = make([]*net.IPNet, 0)
	}
	return &blacklistPlugin{
		blacklist:     blacklist,
		blacklistMask: mask,
	}
}

// blacklistPlugin is a plugin that control only ip addresses in blacklist can **NOT** access services.
type blacklistPlugin struct {
	blacklist     map[string]bool
	blacklistMask []*net.IPNet
}

func (plugin *blacklistPlugin) Connect(conn net.Conn) (net.Conn, bool) {
	ip, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return conn, true
	}
	if plugin.blacklist[ip] {
		return conn, false
	}

	remoteIP := net.ParseIP(ip)
	for _, mask := range plugin.blacklistMask {
		if mask.Contains(remoteIP) {
			return conn, false
		}
	}

	return conn, true
}
