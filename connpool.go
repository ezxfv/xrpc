package xrpc

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"x.io/xrpc/pkg/net"
)

const (
	ActiveHeader = "xrpc/active/"
	poolPort     = 9999
)

type connPool struct {
	addrpool  *sync.Map
	connpool  *sync.Map
	timestamp map[string]time.Time
	cnt       int64
	cc        net.PacketConn
}

func (c *connPool) ActiveOnConn() (conn net.Conn, err error) {
	data := make([]byte, 128, 128)
	n, addr, err := c.cc.ReadFrom(data)
	if err != nil {
		return nil, err
	}
	if n <= len(ActiveHeader) {
		return nil, errors.New("active head is invalid")
	}
	key := string(data[len(ActiveHeader):n])
	list, err := c.getList(addr.String())
	if err != nil {
		return nil, err
	}
	delete(list, key)
	cc, ok := c.connpool.Load(key)
	if !ok {
		c.addrpool.Delete(addr)
		return nil, errors.New("can't take net.Conn")
	}
	return cc.(net.Conn), nil
}

func (c *connPool) getList(addr string) (map[string]bool, error) {
	addrList, ok := c.addrpool.Load(addr)
	if !ok {
		return nil, errors.New("no cached conn")
	}
	list, ok := addrList.(map[string]bool)
	if !ok || len(list) == 0 {
		return nil, errors.New("cached addr list is invalid")
	}
	return list, nil
}

func (c *connPool) Take(addr string) (conn net.Conn, err error) {
	list, err := c.getList(addr)
	if err != nil {
		return nil, err
	}
	var ok bool
	var cc interface{}
	mm := list
	var activeKey string
	for key := range mm {
		cc, ok = c.connpool.Load(key)
		c.connpool.Delete(key)
		delete(mm, key)
		if !ok {
			continue
		}
		conn, ok = cc.(net.Conn)
		if !ok {
			continue
		}
		activeKey = key
		break
	}
	if !ok {
		c.addrpool.Delete(addr)
		return nil, errors.New("cached value isn't a net.Conn")
	}
	c.addrpool.Store(addr, list)
	conn = cc.(net.Conn)
	// 发送个字符串激活连接
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", strings.Split(addr, ":")[0], poolPort))
	if err != nil {
		return nil, err
	}
	n, err := c.cc.WriteTo([]byte(ActiveHeader+activeKey), udpAddr)
	if err != nil {
		return nil, err
	}
	if n != len(ActiveHeader) {
		return nil, errors.New("wrote active header failed")
	}
	return
}

func (c *connPool) Put(conn net.Conn) {
	addr := conn.RemoteAddr().String()
	key := fmt.Sprintf("%s/%d", addr, atomic.AddInt64(&(c.cnt), 1))
	list, _ := c.getList(addr)
	list[key] = true
	c.addrpool.Store(addr, list)
	c.connpool.Store(key, conn)
	c.timestamp[key] = time.Now()
}
