package main

import (
	"flag"
	"net/http"
	"os"

	"x.io/xrpc/pkg/echo"
	"x.io/xrpc/protocol/imdb"
)

var (
	bind     = flag.String("bind", ":8080", "Bind address")
	rootDir  = flag.String("root", "", "Root folder")
	usesGzip = flag.Bool("gzip", true, "Enables gzip/zlib compression")
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
	flag.Parse()
	if len(*rootDir) == 0 {
		*rootDir, _ = os.Getwd()
	}

	e := echo.New()
	e.Debug = true
	e.Use(echo.Logger(), echo.Recovery())

	imdb.RegisterImdbServer("/imdb", e, &ImdbImpl{})
	g := e.Group("/test")
	g.GET("/hello", Hello)

	g1 := e.Group("/file")
	g1.GET("/", index)
	g1.GET("/download", serveFile)
	g1.GET("/download/*path", serveFile)
	g1.HandleFunc("/upload", uploadHandler)

	e.ListenAndServe(*bind)
}
