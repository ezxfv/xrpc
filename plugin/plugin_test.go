package plugin_test

import (
	"math/rand"
	"testing"
	"time"

	"context"

	"x.io/xrpc"
	"x.io/xrpc/plugin"
	"x.io/xrpc/plugin/logp"
	"x.io/xrpc/plugin/prom"
)

func TestLogPlugin(t *testing.T) {
	var (
		logPlugin = logp.New()
		pc        = plugin.NewPluginContainer()
	)
	pc.Add(logPlugin)
	pc.DoPreWriteResponse(nil, nil, nil)
	pc.Remove(logPlugin)
	println()
}

func TestPromPlugin(t *testing.T) {
	var (
		promPlugin = prom.New()
		pc         = plugin.NewPluginContainer()
	)
	pc.Add(promPlugin)
	ctx := context.Background()
	info := &xrpc.UnaryServerInfo{
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
