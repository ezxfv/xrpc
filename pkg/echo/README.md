# Echo
一个基于radix tree的动态路由web框架，Context实现及HTTP常量定义直接引用另一个广为人知的echo(github.com/labstack/echo/v4)。

支持直接注册Handler，注册struct时参考gRPC注册服务的方式生成stub代码，通过接口注释生成路由规则。

## Example
- 定义服务[imdb.go](../../protocol/imdb/imdb.go)
```go
package imdb

import "x.io/xrpc/pkg/echo"

type Imdb interface {
	// GET /findfilm/:name
	FindFilm(c echo.Context) error
	// GET /findfilm/:name/price
	FindFilmPrice(c echo.Context) error
	// GET /findfilm/:name/price/:unit
	FilmPriceUnit(c echo.Context) error
}
```

- 只用parser生成stub代码 [imdb.http.go](../../protocol/imdb/imdb.http.go)
- 实现Imdb接口并注册到echo, stub代码根据接口注释来生成注册规则，启动[echo Server](../../cmd/echo/main.go)
```go
package main

import (
	"net/http"

	"x.io/xrpc/pkg/echo"
	"x.io/xrpc/protocol/imdb"
)

type ImdbImpl struct{}

func (i *ImdbImpl) FindFilm(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "find "+ctx.PathParam("name"))
}

func (i *ImdbImpl) FindFilmPrice(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "The price of "+ctx.PathParam("name")+" is $10")
}

func (i *ImdbImpl) FilmPriceUnit(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "The price of "+ctx.PathParam("name")+" is 10 "+ctx.PathParam("unit"))
}

func Hello(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "hello echo")
}

func main() {
	e := echo.New()
	imdb.RegisterImdbServer("/imdb", e, &ImdbImpl{})
	g := e.Group("/test")
	g.GET("/hello", Hello)
	e.Server(":8080")
}

```