package plugin_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"x.io/xrpc/plugin"
	"x.io/xrpc/plugin/logp"
	"x.io/xrpc/plugin/prom"
	"x.io/xrpc/types"

	_ "x.io/xrpc/plugin/blacklist"
	_ "x.io/xrpc/plugin/chord"
	_ "x.io/xrpc/plugin/crypto"
	_ "x.io/xrpc/plugin/logp"
	_ "x.io/xrpc/plugin/prom"
	_ "x.io/xrpc/plugin/ratelimit"
	_ "x.io/xrpc/plugin/trace"
	_ "x.io/xrpc/plugin/whitelist"
)

func TestLogPlugin(t *testing.T) {
	var (
		logPlugin = logp.New()
		pc        = plugin.NewPluginContainer()
	)
	pc.Add(logPlugin)
	pc.DoPreWriteResponse(nil, nil)
	pc.Remove(logPlugin)
	println()
}

func TestPromPlugin(t *testing.T) {
	var (
		promPlugin = prom.New(nil)
		pc         = plugin.NewPluginContainer()
	)
	pc.Add(promPlugin)
	ctx := context.Background()
	info := &types.UnaryServerInfo{
		Server:     nil,
		FullMethod: "/greeter.Greeter/SayHello",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		time.Sleep(time.Duration(rand.Intn(12)) * time.Millisecond)
		return nil, nil
	}
	for i := 0; i < 100; i++ {
		pc.DoHandle(ctx, nil, info, handler)
	}
}
