package chord

import (
	"encoding/json"
	"net/http"

	"x.io/xrpc/pkg/net"

	echo "x.io/xrpc/pkg/echo"
)

const (
	DefaultAddr = "localhost:9900"
)

type HttpAPI interface {
	set(key, value string)
	get(key string) string
	del(key string)
}

func NewMockChord() *mockChord {
	return &mockChord{
		kvs: map[string]string{},
	}
}

type mockChord struct {
	kvs map[string]string
}

func (m *mockChord) set(key, value string) {
	m.kvs[key] = value
}

func (m *mockChord) get(key string) string {
	return m.kvs[key]
}

func (m *mockChord) del(key string) {
	delete(m.kvs, key)
}

func (m *mockChord) dump(prefix string) {
	d, _ := json.Marshal(m.kvs)
	println(prefix + ": " + string(d))
}

func ServerAPI(addr string, mc HttpAPI) error {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "chord http api")
	})

	g := e.Group("/chord")
	g.POST("/set", func(c echo.Context) error {
		m := make(map[string]string)
		err := c.Bind(&m)
		if err != nil {
			return err
		}
		mc.set(m["key"], m["value"])
		return c.String(http.StatusOK, "")
	})
	g.GET("/get", func(c echo.Context) error {
		key := c.FormValue("key")
		return c.String(http.StatusOK, mc.get(key))
	})
	g.DELETE("/del", func(c echo.Context) error {
		key := c.FormValue("key")
		mc.del(key)
		return c.String(http.StatusOK, "")
	})

	lis, err := net.TCPListen("tcp", addr)
	if err != nil {
		return err
	}
	e.Listener = lis
	return e.Start(addr)
}
