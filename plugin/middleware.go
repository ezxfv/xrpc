package plugin

import (
	"context"

	"github.com/edenzhong7/xrpc/pkg/common"
)

var (
	allBuilders map[string]Builder
)

func init() {
	allBuilders = map[string]Builder{}
}

func RegisterBuilder(m Builder) {
	if _, ok := allBuilders[m.Name()]; !ok {
		allBuilders[m.Name()] = m
	}
}

func PickClientMiddleware(name string) ClientMiddleware {
	return allBuilders[name].NewClientMiddleware()
}

func PickServerMiddleware(name string) ServerMiddleware {
	return allBuilders[name].NewServerMiddleware()
}

func PickSomeClientMiddleware(names []string) []ClientMiddleware {
	var ms []ClientMiddleware
	for _, name := range names {
		if m, ok := allBuilders[name]; ok {
			ms = append(ms, m.NewClientMiddleware())
		}
	}
	return ms
}

func PickSomeServerMiddleware(names []string) []ServerMiddleware {
	var ms []ServerMiddleware
	for _, name := range names {
		if m, ok := allBuilders[name]; ok {
			ms = append(ms, m.NewServerMiddleware())
		}
	}
	return ms
}

type Builder interface {
	Name() string
	NewClientMiddleware() ClientMiddleware
	NewServerMiddleware() ServerMiddleware
}

type ClientMiddleware interface {
	Name() string
	Handle(ctx context.Context, handler interface{}, args ...interface{}) (newHandler interface{})
}

type ServerMiddleware interface {
	Name() string
	Handle(ctx context.Context, handler interface{}, args ...interface{}) (newHandler interface{})
}

func ApplyClientMiddleware(ctx context.Context, ms []ClientMiddleware, handler interface{}, args ...interface{}) (result []interface{}, err error) {
	var newHandler = handler
	for _, m := range ms {
		newHandler = m.Handle(ctx, handler, args...)
	}
	result, err = common.Call(newHandler, newHandler)
	return
}

func ApplyServerMiddleware(ctx context.Context, ms []ServerMiddleware, handler interface{}, args ...interface{}) (result []interface{}, err error) {
	var newHandler = handler
	for _, m := range ms {
		newHandler = m.Handle(ctx, handler, args...)
	}
	result, err = common.Call(newHandler, newHandler)
	return
}
