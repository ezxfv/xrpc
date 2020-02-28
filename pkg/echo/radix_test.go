package echo_test

import (
	"testing"

	"x.io/xrpc/pkg/echo"

	"github.com/stretchr/testify/assert"
)

var (
	nodes = map[string]interface{}{
		"/imdb/findfilm/:name":             "/imdb/findfilm/:name",
		"/imdb/findfilm/:name/price":       "/imdb/findfilm/:name/price",
		"/imdb/findfilm/:name/price/:unit": "/imdb/findfilm/:name/price/:unit",
		"/imdb/v/:op/qq/us/:category":      "/imdb/v/:op/qq/us/:category",
		"/imdb/v/:op/qq/zh/:category/*mp4": "/imdb/v/:op/qq/zh/:category/*mp4",
		"/ftp/*filepath":                   "/ftp/*filepath",
	}
	tree *echo.Tree
)

func init() {
	tree = echo.NewFromMap(nodes)
}

func TestTree(t *testing.T) {
	tree := echo.NewFromMap(nodes)

	v, _, _ := tree.Get("/imdb/findfilm/x/price")
	assert.Equal(t, "/imdb/findfilm/:name/price", v.(string))

	v, _, _ = tree.Get("/imdb/findfilm/xxx")
	assert.Equal(t, "/imdb/findfilm/:name", v.(string))

	v, _, _ = tree.Get("/imdb/findfilm/abc/price/dollar")
	assert.Equal(t, "/imdb/findfilm/:name/price/:unit", v.(string))

	v, _, _ = tree.Get("/imdb/v/list/qq/us/love")
	assert.Equal(t, "/imdb/v/:op/qq/us/:category", v.(string))

	v, _, params := tree.Get("/imdb/v/play/qq/zh/love/2020/x.mp4")
	assert.Equal(t, "/imdb/v/:op/qq/zh/:category/*mp4", v.(string))
	assert.Equal(t, 3, len(params))
	assert.Equal(t, "play", params["op"])
	assert.Equal(t, "love", params["category"])
	assert.Equal(t, "2020/x.mp4", params["mp4"])

	v, _, params = tree.Get("/ftp/films/zh/love/x.mp4")
	assert.Equal(t, "/ftp/*filepath", v.(string))
	assert.Equal(t, "films/zh/love/x.mp4", params["filepath"])
}

func BenchmarkNewRadixTree(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tree.Get("/imdb/findfilm/xxx/price/dollar")
	}
}
