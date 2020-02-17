package config

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var (
	flagKV     = map[string]*Element{}
	envKV      = map[string]*Element{}
	defaultCfg = NewConfig()
)

func NewConfig() *Config {
	cfg := &Config{
		cfgKV:   make(map[string]*Element),
		jsonStr: "",
	}
	return cfg
}

func TimeStamp() string {
	return time.Now().Format("2006-1-2_15:04:05")
}

func SetDefaultKey(k string, v interface{}) {
	envKV[k] = &Element{v}
}

func LoadEnv(keys []string) {
	for _, k := range keys {
		v := os.Getenv(k)
		envKV[k] = &Element{&v}
	}
	return
}

func Parse() {
	flag.Parse()
	args := []*Element{}
	for _, v := range flag.Args() {
		a := v
		args = append(args, &Element{&a})
	}
	flagKV["args"] = &Element{&args}
	return
}

func BoolVar(name string, value bool, usage string) {
	flagKV[name] = &Element{val: &value}
	flag.BoolVar(&value, name, value, usage)
}

func StringVar(name string, value string, usage string) {
	flagKV[name] = &Element{val: &value}
	flag.StringVar(&value, name, value, usage)
}

func IntVar(name string, value int, usage string) {
	flagKV[name] = &Element{val: &value}
	flag.IntVar(&value, name, value, usage)
}

func Int64Var(name string, value int64, usage string) {
	flagKV[name] = &Element{val: &value}
	flag.Int64Var(&value, name, value, usage)
}

func Uint64Var(name string, value uint64, usage string) {
	flagKV[name] = &Element{val: &value}
	flag.Uint64Var(&value, name, value, usage)
}

func Float64Var(name string, value float64, usage string) {
	flagKV[name] = &Element{val: &value}
	flag.Float64Var(&value, name, value, usage)
}

type Element struct {
	val interface{}
}

func (e *Element) Type() string {
	r := reflect.TypeOf(e.val)
	return r.String()
}

func (e *Element) Bool() bool {
	if e.Type() == "bool" {
		return e.val.(bool)
	}
	return false
}

func (e *Element) String() string {
	switch e.val.(type) {
	case *string:
		return *(e.val.(*string))
	}
	return ""
}

func (e *Element) Int() int {
	switch e.val.(type) {
	case *int:
		return *(e.val.(*int))
	case *int32:
		return int(*(e.val.(*int32)))
	case *int64:
		return int(*(e.val.(*int64)))
	case *uint64:
		return int(*(e.val.(*uint64)))
	case *float64:
		return int(*(e.val.(*float64)))
	default:
		return 0
	}
	return *(e.val.(*int))
}

func (e *Element) IntArray() []int {
	v, ok := e.val.(*[]*Element)
	if !ok {
		return []int{}
	}
	vs := make([]int, len(*v), len(*v))
	for i, a := range *v {
		vs[i] = a.Int()
	}
	return vs
}

func (e *Element) IntMap() map[string]int {
	v, ok := e.val.(*map[string]interface{})
	if !ok {
		return nil
	}
	vs := make(map[string]int, len(*v))
	for i, a := range *v {
		vs[i] = a.(*Element).Int()
	}
	return vs
}

func (e *Element) Int64() int64 {
	switch e.val.(type) {
	case *int:
		return int64(*(e.val.(*int)))
	case *int32:
		return int64(*(e.val.(*int32)))
	case *int64:
		return *(e.val.(*int64))
	case *uint64:
		return int64(*(e.val.(*uint64)))
	case *float64:
		return int64(*(e.val.(*float64)))
	}
	return 0
}

func (e *Element) Uint64() uint64 {
	switch e.val.(type) {
	case *int:
		return uint64(*(e.val.(*int)))
	case *int32:
		return uint64(*(e.val.(*int32)))
	case *int64:
		return uint64(*(e.val.(*int64)))
	case *uint64:
		return *(e.val.(*uint64))
	case *float64:
		return uint64(*(e.val.(*float64)))
	}
	return 0
}

func (e *Element) Float() float64 {
	switch e.val.(type) {
	case *int:
		return float64(*(e.val.(*int)))
	case *int32:
		return float64(*(e.val.(*int32)))
	case *int64:
		return float64(*(e.val.(*int64)))
	case *float64:
		return *(e.val.(*float64))
	}
	return 0
}

func (e *Element) FloatArray() []float64 {
	v, ok := e.val.(*[]*Element)
	if !ok {
		return []float64{}
	}
	vs := make([]float64, len(*v), len(*v))
	for i, a := range *v {
		vs[i] = a.Float()
	}
	return vs
}

func (e *Element) FloatMap() map[string]float64 {
	v, ok := e.val.(*map[string]interface{})
	if !ok {
		return nil
	}
	vs := make(map[string]float64, len(*v))
	for i, a := range *v {
		vs[i] = a.(*Element).Float()
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
		vs[i] = a.String()
	}
	return vs
}

func (e *Element) Map() map[string]*Element {
	res := make(map[string]*Element)
	mi := e.val.(*map[string]interface{})
	for k, v := range *mi {
		res[k] = v.(*Element)
	}
	return res
}

func (e *Element) StrMap() map[string]string {
	v, ok := e.val.(*map[string]interface{})
	if !ok {
		return nil
	}
	vs := make(map[string]string, len(*v))
	for i, a := range *v {
		vs[i] = a.(*Element).String()
	}
	return vs
}

type Config struct {
	cfgKV   map[string]*Element
	jsonStr string
	cfgPath string
}

func (c *Config) CfgString() string {
	return c.jsonStr
}

func (c *Config) Dump() error {
	_, err := json.Marshal(&(c.cfgKV))
	if err != nil {
		return nil
	}
	index := strings.LastIndex(c.cfgPath, ".")
	name := c.cfgPath
	ext := ""
	if index > 0 {
		name = c.cfgPath[:index]
		ext = c.cfgPath[index:]
	}
	bakFile := name + "_" + TimeStamp() + ext
	err = os.Rename(c.cfgPath, bakFile)
	if err != nil {
		return err
	}
	f, err := os.Create(c.cfgPath)
	_, err = f.Write([]byte(c.jsonStr))
	return err
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

func (c *Config) LoadCfg(path ...string) error {
	cfgPath := ""
	v, ok := flagKV["config"]
	switch {
	case len(path) > 0:
		cfgPath = path[0]
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
	c.cfgPath = cfgPath
	c.cfgKV["config"] = &Element{&cfgPath}
	c.jsonStr = string(data)
	for k := range m {
		c.cfgKV[k] = c.readCfg(k, gjson.Get(c.jsonStr, k), nil)
	}
	return nil
}

// Ele search level flag>config>env
func (c *Config) Ele(path string) (v *Element, ok bool) {
	if v, ok = flagKV[path]; ok {
		return
	}
	if v, ok = c.cfgKV[path]; ok {
		return
	}
	v, ok = envKV[path]
	return
}

func (c *Config) Set(path string, value interface{}) {
	e := &Element{val: value}
	if _, ok := c.cfgKV[path]; ok {
		flagKV[path] = e
	}
	if _, ok := c.cfgKV[path]; ok {
		c.cfgKV[path] = e
		s, err := sjson.Set(c.jsonStr, path, value)
		if err != nil {
			return
		}
		c.jsonStr = s
	}
	envKV[path] = e
}

func (c *Config) Bool(path string) bool {
	v, ok := c.Ele(path)
	if !ok {
		return false
	}
	return v.Bool()
}

func (c *Config) String(path string) string {
	v, ok := c.Ele(path)
	if !ok {
		return ""
	}
	return v.String()
}

func (c *Config) Int(path string) int {
	v, ok := c.Ele(path)
	if !ok {
		return 0
	}
	return v.Int()
}

func (c *Config) IntArray(path string) []int {
	v, ok := c.Ele(path)
	if !ok {
		return []int{}
	}
	res := v.IntArray()
	return res
}

func (c *Config) IntMap(path string) map[string]int {
	v, ok := c.Ele(path)
	if !ok {
		return nil
	}
	return v.IntMap()
}

func (c *Config) Int64(path string) int64 {
	v, ok := c.Ele(path)
	if !ok {
		return 0
	}
	return v.Int64()
}

func (c *Config) Uint64(path string) uint64 {
	v, ok := c.Ele(path)
	if !ok {
		return 0
	}
	return v.Uint64()
}

func (c *Config) Float(path string) float64 {
	v, ok := c.Ele(path)
	if !ok {
		return 0
	}
	return v.Float()
}

func (c *Config) FloatArray(path string) []float64 {
	v, ok := c.Ele(path)
	if !ok {
		return []float64{}
	}
	return v.FloatArray()
}

func (c *Config) FloatMap(path string) map[string]float64 {
	v, ok := c.Ele(path)
	if !ok {
		return nil
	}
	return v.FloatMap()
}

func (c *Config) Array(path string) []*Element {
	v, ok := c.Ele(path)
	if !ok {
		return []*Element{}
	}
	return v.Array()
}

func (c *Config) StrArray(path string) []string {
	v, ok := c.Ele(path)
	if !ok {
		return []string{}
	}
	return v.StrArray()
}

func (c *Config) Map(path string) map[string]*Element {
	v, ok := c.Ele(path)
	if !ok {
		return nil
	}
	return v.Map()
}

func (c *Config) StrMap(path string) map[string]string {
	v, ok := c.Ele(path)
	if !ok {
		return nil
	}
	return v.StrMap()
}

func (c *Config) Match(pattern string) *Element {
	re := regexp.MustCompile(pattern)
	res := make(map[string]interface{})
	for k, v := range envKV {
		if re.MatchString(k) {
			res[k] = v
		}
	}
	for k, v := range c.cfgKV {
		if re.MatchString(k) {
			res[k] = v
		}
	}
	for k, v := range flagKV {
		if re.MatchString(k) {
			res[k] = v
		}
	}
	return &Element{&res}
}

// Global config
func LoadCfg(path ...string) error {
	return defaultCfg.LoadCfg(path...)
}

// Ele search level flag>config>env
func Ele(path string) (v *Element, ok bool) {
	return defaultCfg.Ele(path)
}

func Set(path string, value interface{}) {
	defaultCfg.Set(path, value)
}

func Bool(path string) bool {
	return defaultCfg.Bool(path)
}

func String(path string) string {
	return defaultCfg.String(path)
}

func Int(path string) int {
	return defaultCfg.Int(path)
}

func IntArray(path string) []int {
	return defaultCfg.IntArray(path)
}

func IntMap(path string) map[string]int {
	return defaultCfg.IntMap(path)
}

func Int64(path string) int64 {
	return defaultCfg.Int64(path)
}

func Uint64(path string) uint64 {
	return defaultCfg.Uint64(path)
}

func Float(path string) float64 {
	return defaultCfg.Float(path)
}

func FloatArray(path string) []float64 {
	return defaultCfg.FloatArray(path)
}

func FloatMap(path string) map[string]float64 {
	return defaultCfg.FloatMap(path)
}

func Array(path string) []*Element {
	return defaultCfg.Array(path)
}

func StrArray(path string) []string {
	return defaultCfg.StrArray(path)
}

func Map(path string) map[string]*Element {
	return defaultCfg.Map(path)
}

func StrMap(path string) map[string]string {
	return defaultCfg.StrMap(path)
}

func Match(pattern string) *Element {
	return defaultCfg.Match(pattern)
}
