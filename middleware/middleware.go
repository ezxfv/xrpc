package middleware

import "context"

type ServerMiddleware interface {
	Handle(ctx context.Context, handler interface{}, args ...interface{})
}

type ClientMiddleware interface {
	Handle(ctx context.Context, handler interface{}, args ...interface{}) (newHandler interface{})
}
