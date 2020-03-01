package echo_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"x.io/xrpc/pkg/echo"

	"github.com/stretchr/testify/assert"
)

var (
	nodes = map[string]interface{}{
		"/imdb/findfilm/:name":                                           "/imdb/findfilm/:name",
		"/imdb/findfilm/:name/price":                                     "/imdb/findfilm/:name/price",
		"/imdb/findfilm/:name/price/:unit":                               "/imdb/findfilm/:name/price/:unit",
		"/imdb/v/:op/qq/us/:category":                                    "/imdb/v/:op/qq/us/:category",
		"/imdb/v/:op/qq/zh/:category/*mp4":                               "/imdb/v/:op/qq/zh/:category/*mp4",
		"/ftp/*filepath":                                                 "/ftp/*filepath",
		"/user/{name:[a-z]{2,10}}/{qq:[1-9][0-9]{4,}}/:sex/:phone/*addr": "regex",
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

	v, _, params = tree.Get("/user/xxx/2521463/male/1234354234/sz/nanshan")
	assert.Equal(t, "regex", v.(string))
	assert.Equal(t, 5, len(params))

	v, _, params = tree.Get("/user/x/100/male/1234354234/sz/nanshan")
	assert.Equal(t, nil, v)
}

func BenchmarkNewRadixTree(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tree.Get("/imdb/findfilm/xxx/price/dollar")
	}
}

func TestRegexp(t *testing.T) {
	s := "/usr/{name:[a-z]{1,10}}/{age:[0-9]{1,2}/:name/:age/*addr"
	p := `(?U)\{(.*)\}`
	re := regexp.MustCompile(p)
	fmt.Printf("%#v\n", re.FindAllStringSubmatch(s, -1))

	p = `(?U)/:(.*)/`
	re = regexp.MustCompile(p)
	if !strings.HasSuffix(s, "/") {
		s += "/"
		s = strings.ReplaceAll(s, "/", "//")
	}
	fmt.Printf("%#v\n", re.FindAllStringSubmatch(s, -1))

	p = `(?U)/\*(.*)/`
	re = regexp.MustCompile(p)
	if !strings.HasSuffix(s, "/") {
		s += "/"
	}
	fmt.Printf("%#v\n", re.FindAllStringSubmatch(s, -1))
}
