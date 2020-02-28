package echo_test

import (
	"testing"

	"x.io/xrpc/pkg/echo"

	"github.com/stretchr/testify/assert"
)

var (
	nodes = map[string]interface{}{
		"/":                                "/",
		"/imdb":                            "imdb",
		"/imdb/findfilm":                   "find film",
		"/imdb/findfilm/:name":             "find film by name",
		"/imdb/findfilm/:name/price":       "get film price",
		"/imdb/findfilm/:name/price/:unit": "get film price unit",
	}
)

func TestTree(t *testing.T) {
	tree := echo.NewFromMap(nodes)
	v, _, _ := tree.Get("/imdb/findfilm/x/price")
	assert.Equal(t, "get film price", v.(string))
	v, _, _ = tree.Get("/imdb/findfilm/xxx")
	assert.Equal(t, "find film by name", v.(string))
	v, _, _ = tree.Get("/imdb/findfilm/xxx/price/dollar")
	assert.Equal(t, "get film price unit", v.(string))
}

func BenchmarkNewRadixTree(b *testing.B) {
	tree := echo.NewFromMap(nodes)
	for i := 0; i < b.N; i++ {
		tree.Get("/imdb/findfilm/xxx/price/dollar")
	}
}
