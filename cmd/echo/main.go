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
	e.ListenAndServe(":8080")
}
