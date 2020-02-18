package config_test

import (
	"bytes"
	"regexp"
	"sort"
	"testing"

	"x.io/xrpc/pkg/config"

	"github.com/stretchr/testify/assert"
)

type Author struct {
	Project string  `yaml:"project,omitempty"`
	Desc    string  `yaml:"desc,omitempty"`
	Version float64 `yaml:"version,omitempty"`
}

type PAuthor struct {
	Author *Author `yaml:"author,omitempty"`
}

func TestYamlParser_Parse(t *testing.T) {
	p := config.NewYamlParser(2)
	p.Parse("./test.yml")
}

func TestYamlParser_Get(t *testing.T) {
	p := config.NewYamlParser(2)
	p.Parse("./test.yml")

	db := p.Get(".dev.db").String()
	assert.Equal(t, "myapp", db)

	f := p.Get(".test.float").Float64()
	assert.Equal(t, 1.11, f)

	project := p.Get(".test.labels.[1].author.project").String()
	assert.Equal(t, "xrpc", project)

	w := bytes.NewBuffer([]byte(nil))
	p.Get(".test.labels.[1]").Dump(w)
	println(w.String())
}

func TestYamlNode_Marshal(t *testing.T) {
	p := config.NewYamlParser(2)
	p.Parse("./test.yml")

	a := Author{}
	err := p.Get(".test.labels.[1].author").Marshal(&a)
	assert.Equal(t, nil, err)
	assert.Equal(t, "xrpc", a.Project)
	assert.Equal(t, "a simple rpc framework", a.Desc)
	assert.Equal(t, 1.0, a.Version)

	pa := PAuthor{}
	err = p.Get(".test.labels.[1]").Marshal(&pa)
	assert.Equal(t, nil, err)
	assert.Equal(t, "xrpc", pa.Author.Project)
	assert.Equal(t, "a simple rpc framework", pa.Author.Desc)
	assert.Equal(t, 1.0, pa.Author.Version)
}

func TestYamlNode_Dump(t *testing.T) {
	p := config.NewYamlParser(2)
	p.Parse("./test.yml")

	w := bytes.NewBuffer([]byte(nil))
	p.Get(".test.ints").Dump(w)
	println(w.String())

	w = bytes.NewBuffer([]byte(nil))
	p.Get(".test.host").Dump(w)
	println(w.String())

	w = bytes.NewBuffer([]byte(nil))
	f := p.Get(".test.float")
	f.Value = "2.0"
	f.Dump(w)
	println(w.String())

	w = bytes.NewBuffer([]byte(nil))
	p.Get(".test.labels.[1]").Dump(w)
	println(w.String())

	w = bytes.NewBuffer([]byte(nil))
	p.Get(".test.labels.[0].type").Dump(w)
	println(w.String())

	w = bytes.NewBuffer([]byte(nil))
	p.Get(".test.labels").Dump(w)
	println(w.String())

	w = bytes.NewBuffer([]byte(nil))
	p.Get(".test").Dump(w)
	println(w.String())

	w = bytes.NewBuffer([]byte(nil))
	p.Get(".").Dump(w)
	println(w.String())
}

func TestArray(t *testing.T) {
	p := config.NewYamlParser(2)
	p.Parse("./test.yml")

	ints := p.Get(".test.ints").IntArray()
	assert.Equal(t, []int{1, 2, 3}, ints)

	fs := p.Get(".test.fs").FloatArray()
	assert.Equal(t, []float64{1.1, 2.2, 3.3}, fs)

	ss := p.Get(".test.ss").StrArray()
	assert.Equal(t, []string{"a", "b", "c"}, ss)
}

func TestAlias(t *testing.T) {
	p := config.NewYamlParser(2)
	p.Parse("./test.yml")

	host := p.Get(".test.host").String()
	assert.Equal(t, "localhost", host)
}

func TestChainGet(t *testing.T) {
	p := config.NewYamlParser(2)
	p.Parse("./test.yml")

	host := p.Get(".test").Get(".host").String()
	assert.Equal(t, "localhost", host)

	webs := p.Get(".test.labels.[1]").Get(".author.web").StrArray()
	assert.Equal(t, []string{"baidu.com", "google.com"}, webs)
}

func TestSquareSymbol(t *testing.T) {
	p := config.NewYamlParser(2)
	p.Parse("./test.yml")

	webs := p.Get(".test.labels.[1].author.web").StrArray()
	assert.Equal(t, []string{"baidu.com", "google.com"}, webs)
}

func BenchmarkArray(b *testing.B) {
	p := config.NewYamlParser(2)
	p.Parse("./test.yml")
	for i := 0; i < b.N; i++ {
		webs := p.Get(".test.labels.[1].author.web").StrArray()
		assert.Equal(b, []string{"baidu.com", "google.com"}, webs)
	}
}

func TestYamlNode_Match(t *testing.T) {
	p := config.NewYamlParser(2)
	p.Parse("./test.yml")

	keys := p.Match(".test.maps.[*].key").StrArray()
	sort.Strings(keys)
	assert.Equal(t, []string{"k1", "k2", "k3"}, keys)

	keys = p.Match(".test.maps.[^1].key").StrArray()
	sort.Strings(keys)
	assert.Equal(t, []string{"k1", "k3"}, keys)

	keys = p.Match(".test.maps.[1-2].key").StrArray()
	sort.Strings(keys)
	assert.Equal(t, []string{"k2", "k3"}, keys)

	keys = p.Match("(.*).[^3].key").StrArray()
	sort.Strings(keys)
	assert.Equal(t, []string{"k1", "k2", "k3"}, keys)

	fs := p.Match("(.*).fs.[^1]").FloatArray()
	sort.Float64s(fs)
	assert.Equal(t, []float64{1.1, 3.3}, fs)
}

func TestGenRegexp(t *testing.T) {
	pattern := ".test.maps.[*].key"
	src0 := ".test.maps.[0].key"
	src1 := ".test.maps.[1].key"
	src2 := ".test.maps.[2].key"
	pattern = config.GenStdRegexp(pattern)
	r := regexp.MustCompile(pattern)
	assert.Equal(t, true, r.MatchString(src0))

	pattern = ".test.maps.[1-2].key"
	pattern = config.GenStdRegexp(pattern)
	r = regexp.MustCompile(pattern)
	assert.Equal(t, true, r.MatchString(src1))
	assert.Equal(t, true, r.MatchString(src2))

	pattern = ".test.maps.[^1].key"
	pattern = config.GenStdRegexp(pattern)
	r = regexp.MustCompile(pattern)
	assert.Equal(t, true, r.MatchString(src0))
	assert.Equal(t, true, r.MatchString(src2))
}
