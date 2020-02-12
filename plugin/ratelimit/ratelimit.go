package ratelimit

import (
	"time"

	"x.io/xrpc/pkg/net"

	"github.com/juju/ratelimit"
)

// NewRateLimitingPlugin creates a new RateLimitingPlugin
func New(fillInterval time.Duration, capacity int64) *RateLimitingPlugin {
	tb := ratelimit.NewBucket(fillInterval, capacity)

	return &RateLimitingPlugin{
		FillInterval: fillInterval,
		Capacity:     capacity,
		bucket:       tb,
	}
}

// RateLimitingPlugin can limit connecting per unit time
type RateLimitingPlugin struct {
	FillInterval time.Duration
	Capacity     int64
	bucket       *ratelimit.Bucket
}

// HandleConnAccept can limit connecting rate
func (plugin *RateLimitingPlugin) Connect(conn net.Conn) (net.Conn, bool) {
	return conn, plugin.bucket.TakeAvailable(1) > 0
}
