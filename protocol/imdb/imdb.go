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
