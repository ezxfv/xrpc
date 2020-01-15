package xrpc

import (
	_ "github.com/bluele/gcache"
	_ "github.com/google/uuid"
	_ "github.com/gorilla/websocket"
	_ "github.com/kavu/go_reuseport"
	_ "github.com/klauspost/reedsolomon"
	_ "github.com/lucas-clemente/quic-go"
	_ "github.com/prometheus/client_golang/prometheus"
	_ "github.com/soheilhy/cmux"
	_ "github.com/stretchr/testify"
	_ "github.com/xtaci/kcp-go"
	_ "github.com/xtaci/smux"
)

const (
	PkgVersion               string = "0.0.1"
	SupportPackageIsVersion4 int    = 4
)
