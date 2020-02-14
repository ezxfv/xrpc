package chord

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"x.io/xrpc"
	"x.io/xrpc/api"
	"x.io/xrpc/pkg/net"
	chordp "x.io/xrpc/plugin/chord"
	cryptop "x.io/xrpc/plugin/crypto"
	"x.io/xrpc/plugin/prom"
	"x.io/xrpc/plugin/trace"
	"x.io/xrpc/plugin/whitelist"
	"x.io/xrpc/protocol/chordpb"
)

var (
	enablePlugin = true
	enableTrace  = false
	enableAuth   = false
	enableCrypto = false
	enableAPI    = false
	enableChord  = true

	sessionID  = "session_math_0"
	sessionKey = "1234"

	user = "admin"
	pass = "1234"
)

func parseAddr(addr string) (host string, port int) {
	arr := strings.Split(addr, ":")
	host = arr[0]
	port, _ = strconv.Atoi(arr[1])
	return
}

func newServer(protocol, addr string) (lis net.Listener, svr *xrpc.Server) {
	lis, err := net.Listen(context.Background(), protocol, addr)
	if err != nil {
		log.Fatal(err)
	}
	s := xrpc.NewServer()
	if enablePlugin {
		promPlugin := prom.New()
		//logPlugin := logp.New()
		//promPlugin.Collect(logPlugin.Logger().EnableCounter())
		whitelistPlugin := whitelist.New(map[string]bool{"127.0.0.1": true}, nil)
		s.ApplyPlugins(promPlugin, whitelistPlugin)
	}
	if enableTrace {
		tracePlugin := trace.New()
		s.ApplyPlugins(tracePlugin)
	}
	if enableCrypto {
		cryptoPlugin := cryptop.New()
		cryptoPlugin.SetKey(sessionID, sessionKey)
		s.ApplyPlugins(cryptoPlugin)
	}
	if enableChord {
		chordPlugin := chordp.New(fmt.Sprintf("%s://%s", protocol, addr), "http://localhost:9900/chord")
		s.ApplyPlugins(chordPlugin)
	}
	s.StartPlugins()
	if enableAuth {
		admin := xrpc.NewAdminAuthenticator(user, pass)
		s.SetAuthenticator(admin)
	}
	if enableAPI {
		go api.Server(":8080")
	}
	return lis, s
}

func (c *chordImpl) Server() error {
	lis, s := newServer("tcp", fmt.Sprintf("%s:%d", c.host, c.port))
	chordpb.RegisterChordServer(s, c)
	go c.stabilize()
	return s.Serve(lis)
}
