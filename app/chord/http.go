package chord

import (
	"encoding/json"
	"net/http"

	echo "github.com/labstack/echo/v4"
	"x.io/xrpc/pkg/net"

	"github.com/labstack/echo/v4/middleware"
)

const (
	DefaultAddr = "localhost:9900"
)

func newMockChord() *mockChord {
	return &mockChord{
		kvs: map[string]string{},
	}
}

type mockChord struct {
	kvs map[string]string
}

func (m *mockChord) Set(key, value string) {
	m.kvs[key] = value
}

func (m *mockChord) Get(key string) (value string) {
	return m.kvs[key]
}

func (m *mockChord) Del(key string) {
	delete(m.kvs, key)
}

func (m *mockChord) dump(prefix string) {
	d, _ := json.Marshal(m.kvs)
	println(prefix + ": " + string(d))
}

func Server(addr string) error {
	e := echo.New()
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	mc := newMockChord()

	// Routes
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
		mc.Set(m["key"], m["value"])
		mc.dump("set")
		return c.String(http.StatusOK, "")
	})
	g.GET("/get", func(c echo.Context) error {
		key := c.FormValue("key")
		mc.dump("get")
		return c.String(http.StatusOK, mc.Get(key))
	})
	g.DELETE("/del", func(c echo.Context) error {
		key := c.FormValue("key")
		mc.Del(key)
		mc.dump("del")
		return c.String(http.StatusOK, "")
	})

	lis, err := net.TCPListen("tcp", addr)
	if err != nil {
		return err
	}
	e.Listener = lis
	// Start serve
	return e.Start(addr)
}
