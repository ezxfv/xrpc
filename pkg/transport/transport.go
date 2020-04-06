package transport

import (
	"context"
)

type Transport interface {
	Protocol() string
	SendMsg(ctx context.Context, m interface{}) error
	RecvMsg(ctx context.Context, m interface{}) (context.Context, error)
	Close() error
}
