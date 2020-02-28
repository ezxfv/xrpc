package api

import (
	"net/http"

	"x.io/xrpc/pkg/net"

	echo "x.io/xrpc/pkg/echo"
)

var (
	e *echo.Echo
)

func init() {
	e = echo.New()
}

type APIer interface {
	RegisterAPI(e *echo.Echo)
}

func Register(api APIer) {
	api.RegisterAPI(e)
}

func Server(addr string) error {
	// Routes
	e.GET("/", hello)

	lis, err := net.TCPListen("tcp", addr)
	if err != nil {
		return err
	}
	e.Listener = lis
	// Start serve
	return e.Start(addr)
}

// Handler
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "XRPC api is working!")
}
