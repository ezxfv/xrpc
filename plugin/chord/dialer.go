package chord

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	chord "x.io/xrpc/app/chord/client"

	"x.io/xrpc/pkg/net"
)

func init() {
	net.RegisterDialer("chord", chordDial)
}

// chord://math.Math 或者 chord://custom.math
func chordDial(ctx context.Context, addr string) (net.Conn, error) {
	// 解析chord地址获取对应服务的信息，再根据对应的协议Dial后返回连接
	c := chord.NewChordClient(chord.DefaultURL)
	serviceJson, err := c.Get(addr)
	if err != nil {
		return nil, err
	}
	if len(serviceJson) == 0 {
		return nil, errors.New("chord dial error: no such service " + addr)
	}
	s := &service{}
	err = json.Unmarshal([]byte(serviceJson), s)
	if err != nil {
		return nil, err
	}
	if len(s.Endpoints) == 0 {
		return nil, errors.New("no valuable addr for service " + addr)
	}
	for k := range s.Endpoints {
		arr := strings.Split(k, "://")
		if len(arr) != 2 {
			continue
		}
		conn, err := net.Dial(ctx, arr[0], arr[1])
		if err != nil {
			continue
		}
		println("chord: connected to " + k)
		return conn, nil
	}
	return nil, errors.New("parse all addresses failed for service " + addr)
}
