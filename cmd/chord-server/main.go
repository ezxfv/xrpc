package main

import (
	"context"
	"fmt"
	"log"

	"x.io/xrpc"
	"x.io/xrpc/api"
	"x.io/xrpc/app/chord"
	"x.io/xrpc/pkg/net"
	chordp "x.io/xrpc/plugin/chord"
	"x.io/xrpc/plugin/crypto"
	"x.io/xrpc/plugin/logp"
	"x.io/xrpc/plugin/prom"
	"x.io/xrpc/plugin/trace"
	"x.io/xrpc/plugin/whitelist"
	"x.io/xrpc/protocol/chordpb"
)

var (
	enablePlugin = true
	enableAuth   = true
	enableCrypto = true
	enableAPI    = true
	enableChord  = true

	sessionID  = "session_math_0"
	sessionKey = "1234"

	user = "admin"
	pass = "1234"
)

func newServer(protocol, addr string) (lis net.Listener, svr *xrpc.Server) {
	lis, err := net.Listen(context.Background(), protocol, addr)
	if err != nil {
		log.Fatal(err)
	}
	s := xrpc.NewServer()
	if enablePlugin {
		promPlugin := prom.New()
		logPlugin := logp.New()
		tracePlugin := trace.New()
		whitelistPlugin := whitelist.New(map[string]bool{"127.0.0.1": true}, nil)
		s.ApplyPlugins(promPlugin, tracePlugin, whitelistPlugin, logPlugin)
		if enableCrypto {
			cryptoPlugin := crypto.New()
			cryptoPlugin.SetKey(sessionID, sessionKey)
			s.ApplyPlugins(cryptoPlugin)
		}
		if enableChord {
			chordPlugin := chordp.New(fmt.Sprintf("%s://%s", protocol, addr), "http://localhost:9900/chord")
			s.ApplyPlugins(chordPlugin)
		}
		s.StartPlugins()
	}
	if enableAuth {
		admin := xrpc.NewAdminAuthenticator(user, pass)
		s.SetAuthenticator(admin)
	}
	if enableAPI {
		go api.Server(":8080")
	}
	return lis, s
}

func main() {
	c := chord.NewChord("localhost", 9080, nil, nil)
	lis, s := newServer("tcp", "localhost:9080")
	chordpb.RegisterChordServer(s, c)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	go chord.Server(chord.DefaultAddr, chord.NewChordAPI(c))
	s.Start()
}
