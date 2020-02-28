package echo_test

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"testing"

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

func TestEcho(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	e := echo.New()
	e.Debug = true
	e.Cache(true)

	e.Use(echo.Recovery())
	e.EnableListRoutes()

	imdb.RegisterImdbServer("/imdb", e, &ImdbImpl{})
	g := e.Group("/test")
	g.GET("/hello", Hello)
	go e.ListenAndServe(":8080")
	log.Println(http.ListenAndServe("localhost:3999", nil))
}

func BenchmarkEcho_GET(b *testing.B) {
	for i := 0; i < b.N; i++ {
		http.Get("http://localhost:8080/imdb/findfilm/x/price/dollar")
	}
}
