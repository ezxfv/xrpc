package config

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strconv"

	"github.com/tidwall/gjson"
)

var (
	CfgPath = "./config.json"
)

func DefaultConfig() *Config {
	cfg := NewConfig()
	return cfg
}

type Element struct {
	val interface{}
}

func (e *Element) Type() string {
	r := reflect.TypeOf(e.val)
	return r.Name()
}

func (e *Element) String() string {
	return *(e.val.(*string))
}

func (e *Element) Int() int {
	return *(e.val.(*int))
}

func (e *Element) IntArray() []int {
	v, ok := e.val.(*[]*Element)
	if !ok {
		return []int{}
	}
	vs := make([]int, len(*v), len(*v))
	for i, a := range *v {
		vs[i] = *(a.val.(*int))
	}
	return vs
}

func (e *Element) IntMap() map[string]int {
	v, ok := e.val.(map[string]*Element)
	if !ok {
		return nil
	}
	vs := make(map[string]int, len(v))
	for i, a := range v {
		vs[i] = int(*(a.val.(*float64)))
	}
	return vs
}

func (e *Element) Int64() int64 {
	return *(e.val.(*int64))
}

func (e *Element) UInt64() uint64 {
	return *(e.val.(*uint64))
}

func (e *Element) Float() float64 {
	return *(e.val.(*float64))
}

func (e *Element) FloatArray() []float64 {
	v, ok := e.val.(*[]*Element)
	if !ok {
		return []float64{}
	}
	vs := make([]float64, len(*v), len(*v))
	for i, a := range *v {
		vs[i] = *(a.val.(*float64))
	}
	return vs
}

func (e *Element) FloatMap() map[string]float64 {
	v, ok := e.val.(map[string]*Element)
	if !ok {
		return nil
	}
	vs := make(map[string]float64, len(v))
	for i, a := range v {
		vs[i] = *(a.val.(*float64))
	}
	return vs
}

func (e *Element) Array() []*Element {
	return e.val.([]*Element)
}

func (e *Element) StrArray() []string {
	v, ok := e.val.([]*Element)
	if !ok {
		return []string{}
	}
	vs := make([]string, len(v), len(v))
	for i, a := range v {
		vs[i] = *(a.val.(*string))
	}
	return vs
}

func (e *Element) Map() map[string]*Element {
	return e.val.(map[string]*Element)
}

func (e *Element) StrMap() map[string]string {
	v, ok := e.val.(map[string]*Element)
	if !ok {
		return nil
	}
	vs := make(map[string]string, len(v))
	for i, a := range v {
		vs[i] = *(a.val.(*string))
	}
	return vs
}

func NewConfig() *Config {
	cfg := &Config{
		cfgKV:   make(map[string]*Element),
		flagKV:  make(map[string]*Element),
		envKV:   make(map[string]*Element),
		jsonStr: "",
	}
	return cfg
}

type Config struct {
	cfgKV   map[string]*Element
	flagKV  map[string]*Element
	envKV   map[string]*Element
	jsonStr string
}

func (c *Config) Parse() {
	flag.Parse()
	return
}

func (c *Config) readCfg(path string, res gjson.Result, pre *Element) *Element {
	switch res.Type {
	case gjson.String:
		v := res.String()
		e := &Element{&v}
		c.cfgKV[path] = e
		return e
	case gjson.Number:
		switch {
		case res.Index > 0:
			v := res.Index
			e := &Element{&v}
			c.cfgKV[path] = e
			return e
		case res.Num > 0:
			v := res.Num
			e := &Element{&v}
			c.cfgKV[path] = e
			return e
		default:
			v := 0
			e := &Element{&v}
			c.cfgKV[path] = e
			return e
		}
	case gjson.JSON:
		switch {
		case len(res.Array()) > 0:
			local := make([]*Element, len(res.Array()))
			for i, v := range res.Array() {
				local[i] = c.readCfg(path+"."+strconv.Itoa(i), v, nil)
			}
			c.cfgKV[path] = &Element{&local}
			var ele []*Element
			if pre != nil {
				ele = append(pre.val.([]*Element), local...)
			} else {
				ele = local
			}
			return &Element{&ele}
		case len(res.Map()) > 0:
			local := make(map[string]interface{})
			for i, v := range res.Map() {
				local[i] = c.readCfg(path+"."+i, v, nil)
			}
			e := &Element{&local}
			c.cfgKV[path] = e
			return e
		default:
			v := map[string]interface{}{}
			e := &Element{&v}
			c.cfgKV[path] = e
			return e
		}
	}
	v := "null"
	e := &Element{&v}
	c.cfgKV[path] = e
	return e
}

func (c *Config) LoadCfg() error {
	cfgPath := ""
	v, ok := c.flagKV["config"]
	switch {
	case CfgPath != "":
		cfgPath = CfgPath
	case ok && v != nil:
		cfgPath = v.String()
	default:
		cfgPath = "./config.json"
	}
	f, err := os.Open(cfgPath)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	m := make(map[string]interface{})
	err = json.Unmarshal(data, &m)
	if err != nil {
		return err
	}
	c.jsonStr = string(data)
	for k := range m {
		c.cfgKV[k] = c.readCfg(k, gjson.Get(c.jsonStr, k), nil)
	}
	return nil
}

func (c *Config) LoadEnv(keys []string) {
	for _, k := range keys {
		v := os.Getenv(k)
		c.envKV[k] = &Element{&v}
	}
	return
}

// Get search level flag>config>env
func (c *Config) Get(path string) (v *Element, ok bool) {
	if v, ok = c.flagKV[path]; ok {
		return
	}
	if v, ok = c.cfgKV[path]; ok {
		return
	}
	v, ok = c.envKV[path]
	return
}

func (c *Config) Set(path string, value interface{}) {
	c.cfgKV[path] = &Element{val: value}
}

func (c *Config) GetStr(path string) string {
	v, ok := c.Get(path)
	if !ok {
		return ""
	}
	return v.String()
}

func (c *Config) GetInt(path string) int {
	v, ok := c.Get(path)
	if !ok {
		return 0
	}
	return v.Int()
}

func (c *Config) GetIntArray(path string) []int {
	v, ok := c.Get(path)
	if !ok {
		return []int{}
	}
	res := v.IntArray()
	return res
}

func (c *Config) GetIntMap(path string) map[string]int {
	v, ok := c.Get(path)
	if !ok {
		return nil
	}
	return v.IntMap()
}

func (c *Config) GetInt64(path string) int64 {
	v, ok := c.Get(path)
	if !ok {
		return 0
	}
	return v.Int64()
}

func (c *Config) GetUInt64(path string) uint64 {
	v, ok := c.Get(path)
	if !ok {
		return 0
	}
	return v.UInt64()
}

func (c *Config) GetFloat(path string) float64 {
	v, ok := c.Get(path)
	if !ok {
		return 0
	}
	return v.Float()
}

func (c *Config) GetFloatArray(path string) []float64 {
	v, ok := c.Get(path)
	if !ok {
		return []float64{}
	}
	return v.FloatArray()
}

func (c *Config) GetFloatMap(path string) map[string]float64 {
	v, ok := c.Get(path)
	if !ok {
		return nil
	}
	return v.FloatMap()
}

func (c *Config) GetArray(path string) []*Element {
	v, ok := c.Get(path)
	if !ok {
		return []*Element{}
	}
	return v.Array()
}

func (c *Config) GetStrArray(path string) []string {
	v, ok := c.Get(path)
	if !ok {
		return []string{}
	}
	return v.StrArray()
}

func (c *Config) GetMap(path string) map[string]*Element {
	v, ok := c.Get(path)
	if !ok {
		return nil
	}
	return v.Map()
}

func (c *Config) GetStrMap(path string) map[string]string {
	v, ok := c.Get(path)
	if !ok {
		return nil
	}
	return v.StrMap()
}

func (c *Config) String(name string, value string, usage string) {
	var v string
	c.flagKV[name] = &Element{val: &v}
	flag.StringVar(&v, name, value, usage)
}

func (c *Config) Int(name string, value int, usage string) {
	var v int
	c.flagKV[name] = &Element{val: &v}
	flag.IntVar(&v, name, value, usage)
}

func (c *Config) Int64(name string, value int64, usage string) {
	var v int64
	c.flagKV[name] = &Element{val: &v}
	flag.Int64Var(&v, name, value, usage)
}

func (c *Config) Uint64(name string, value uint64, usage string) {
	var v uint64
	c.flagKV[name] = &Element{val: &v}
	flag.Uint64Var(&v, name, value, usage)
}

func (c *Config) Float64(name string, value float64, usage string) {
	var v float64
	c.flagKV[name] = &Element{val: &v}
	flag.Float64Var(&v, name, value, usage)
}

func (c *Config) Match(pattern string) *Element {
	re := regexp.MustCompile(pattern)
	res := make(map[string]*Element)
	for k, v := range c.envKV {
		if re.MatchString(k) {
			res[k] = v
		}
	}
	for k, v := range c.cfgKV {
		if re.MatchString(k) {
			res[k] = v
		}
	}
	for k, v := range c.flagKV {
		if re.MatchString(k) {
			res[k] = v
		}
	}
	return &Element{res}
}
