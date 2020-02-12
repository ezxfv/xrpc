package whitelist

import "x.io/xrpc/pkg/net"

func New(whitelist map[string]bool, mask []*net.IPNet) *whitelistPlugin {
	if whitelist == nil {
		whitelist = make(map[string]bool)
	}
	if mask == nil {
		mask = make([]*net.IPNet, 0)
	}
	return &whitelistPlugin{
		whitelist:     whitelist,
		whitelistMask: mask,
	}
}

type whitelistPlugin struct {
	whitelist     map[string]bool
	whitelistMask []*net.IPNet
}

func (plugin *whitelistPlugin) Connect(conn net.Conn) (net.Conn, bool) {
	ip, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return conn, false
	}

	if plugin.whitelist[ip] {
		return conn, true
	}

	remoteIP := net.ParseIP(ip)
	for _, mask := range plugin.whitelistMask {
		if mask.Contains(remoteIP) {
			return conn, true
		}
	}

	return conn, false
}
