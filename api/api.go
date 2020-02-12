package api

import (
	"net/http"

	"x.io/xrpc/pkg/net"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

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
