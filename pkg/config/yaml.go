package config

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v2"
)

type ValueType int

const (
	IntVal ValueType = iota
	BoolVal
	Float64Val
	StringVal
	ArrayVal
	MapVal
)

type YamlNode struct {
	Name  string
	Path  string
	Ident int
	Depth int
	Line  int
	Value string

	ArrNodes []*YamlNode
	MapNodes []*YamlNode

	prev      *YamlNode
	alias     map[string]string
	aliasNode map[string]*YamlNode
}

func (n *YamlNode) Marshal(v interface{}) (err error) {
	w := bytes.NewBuffer([]byte(nil))
	n.Dump(w)
	println(w.String())
	return yaml.Unmarshal(w.Bytes(), v)
}

func (n *YamlNode) addAliasNode(name string, a *YamlNode) {
	if n.aliasNode == nil {
		n.aliasNode = map[string]*YamlNode{}
	}
	n.aliasNode[name] = a
}

func (n *YamlNode) value(t ValueType) interface{} {
	switch t {
	case StringVal:
		return n.Value
	case BoolVal:
		return len(n.Value) > 0 && n.Value == "true"
	case IntVal:
		i, _ := strconv.Atoi(n.Value)
		return i
	case Float64Val:
		f, _ := strconv.ParseFloat(n.Value, 10)
		return f
	case ArrayVal:
		return n.ArrNodes
	case MapVal:
		m := map[string]*YamlNode{}
		for _, mn := range n.MapNodes {
			m[mn.Name] = mn
		}
		return m
	}
	return nil
}

func (n *YamlNode) Get(path string) *YamlNode {
	if !strings.HasPrefix(path, n.Path) {
		path = n.Path + path
	}
	if newPath, ok := n.alias[path]; ok {
		path = newPath
	}
	for _, c := range n.MapNodes {
		if c.Path == path {
			return c
		}
		if strings.HasPrefix(path, c.Path) {
			return c.Get(path)
		}
	}
	for _, c := range n.ArrNodes {
		if c.Path == path {
			return c
		}
		if strings.HasPrefix(path, c.Path) {
			return c.Get(path)
		}
	}
	for _, c := range n.aliasNode {
		if c.Path == path {
			return c
		}
		if strings.HasPrefix(path, c.Path) {
			return c.Get(path)
		}
	}
	return &YamlNode{}
}

func (n *YamlNode) String() string {
	if len(n.Value) == 0 {
		return ""
	}
	return n.value(StringVal).(string)
}

func (n *YamlNode) Bool() bool {
	if len(n.Value) == 0 {
		return false
	}
	return n.value(BoolVal).(bool)
}

func (n *YamlNode) Int() int {
	if len(n.Value) == 0 {
		return 0
	}
	return n.value(IntVal).(int)
}

func (n *YamlNode) Float64() float64 {
	if len(n.Value) == 0 {
		return 0
	}
	return n.value(Float64Val).(float64)
}

func (n *YamlNode) Map() map[string]*YamlNode {
	if len(n.Value) == 0 {
		return map[string]*YamlNode{}
	}
	return n.value(MapVal).(map[string]*YamlNode)
}

func (n *YamlNode) Array() []*YamlNode {
	if len(n.Value) == 0 {
		return []*YamlNode{}
	}
	return n.value(ArrayVal).([]*YamlNode)
}

func (n *YamlNode) BoolMap() map[string]bool {
	if len(n.ArrNodes) == 0 {
		return map[string]bool{}
	}
	is := n.value(ArrayVal).([]*YamlNode)
	res := map[string]bool{}
	for _, s := range is {
		res[s.Name] = s.Bool()
	}
	return res
}

func (n *YamlNode) BoolArray() []bool {
	if len(n.ArrNodes) == 0 {
		return []bool{}
	}
	is := n.value(ArrayVal).([]*YamlNode)
	var res []bool
	for _, s := range is {
		res = append(res, s.Bool())
	}
	return res
}

func (n *YamlNode) IntMap() map[string]int {
	if len(n.ArrNodes) == 0 {
		return map[string]int{}
	}
	is := n.value(ArrayVal).([]*YamlNode)
	res := map[string]int{}
	for _, s := range is {
		res[s.Name] = s.Int()
	}
	return res
}

func (n *YamlNode) IntArray() []int {
	if len(n.ArrNodes) == 0 {
		return []int{}
	}
	is := n.value(ArrayVal).([]*YamlNode)
	var res []int
	for _, s := range is {
		res = append(res, s.Int())
	}
	return res
}

func (n *YamlNode) FloatMap() map[string]float64 {
	if len(n.ArrNodes) == 0 {
		return map[string]float64{}
	}
	is := n.value(ArrayVal).([]*YamlNode)
	res := map[string]float64{}
	for _, s := range is {
		res[s.Name] = s.Float64()
	}
	return res
}

func (n *YamlNode) FloatArray() []float64 {
	if len(n.ArrNodes) == 0 {
		return []float64{}
	}
	is := n.value(ArrayVal).([]*YamlNode)
	var res []float64
	for _, s := range is {
		res = append(res, s.Float64())
	}
	return res
}

func (n *YamlNode) StrMap() map[string]string {
	if len(n.ArrNodes) == 0 {
		return map[string]string{}
	}
	is := n.value(ArrayVal).([]*YamlNode)
	res := map[string]string{}
	for _, s := range is {
		res[s.Name] = s.String()
	}
	return res
}

func (n *YamlNode) StrArray() []string {
	if len(n.ArrNodes) == 0 {
		return []string{}
	}
	is := n.value(ArrayVal).([]*YamlNode)
	var res []string
	for _, s := range is {
		res = append(res, s.String())
	}
	return res
}

var printf = func(format string, args ...interface{}) []byte {
	return []byte(fmt.Sprintf(format, args...))
}

func genSpace(ident, depth int) []byte {
	n := ident * depth
	s := ""
	for i := 0; i < n; i++ {
		s += " "
	}
	return []byte(s)
}

func (n *YamlNode) Dump(w *bytes.Buffer, start ...int) {
	s := n.Depth
	if len(start) > 0 {
		s = start[0]
	}
	for aliasName, c := range n.aliasNode {
		if strings.Contains(w.String(), aliasName) {
			continue
		}
		lw := bytes.NewBuffer([]byte(nil))
		c.Dump(lw, 0)
		lw.Write(w.Bytes())
		w.Reset()
		w.Write(lw.Bytes())
	}
	name := n.Name
	depth := n.Depth - s - 1
	if depth < 0 {
		depth = 0
	}
	space := genSpace(n.Ident, depth)
	if len(n.Value) > 0 {
		w.Write(space)
		if isElem(n.Name) {
			w.Write(printf("- %s\n", n.Value))
		} else {
			if isElem(n.prev.Name) {
				w.Write([]byte("- "))
			}
			w.Write(printf("%s: %s\n", name, n.Value))
		}
		if !strings.HasPrefix(n.Value, "&") {
			return
		}
	}

	if len(n.ArrNodes) > 0 {
		if len(start) == 1 && !isElem(n.Name) {
			w.Write(printf("%s%s:\n", space, n.Name))
		}
		for _, an := range n.ArrNodes {
			an.Dump(w, s)
		}
		return
	}
	if len(n.MapNodes) > 0 {
		if len(start) == 1 && !isElem(n.Name) {
			if isElem(n.prev.Name) && s != n.prev.Depth {
				if len(space) >= 2 {
					space = space[2:]
				}
				w.Write(printf("%s- %s:\n", space, n.Name))
			} else if len(n.Value) == 0 {
				w.Write(printf("%s%s:\n", space, n.Name))
			}
		}
		for _, mn := range n.MapNodes {
			mn.Dump(w, s)
		}
	}
}

func NewYamlParser(ident int) *YamlParser {
	return &YamlParser{
		Ident: ident,
		alias: map[string]*YamlNode{},
		root: &YamlNode{
			Depth: 0,
			Ident: ident,
		},
		kvs: map[string]*YamlNode{},
	}
}

type YamlParser struct {
	Ident int
	alias map[string]*YamlNode
	root  *YamlNode

	kvs map[string]*YamlNode
}

func (y *YamlParser) Get(path string) *YamlNode {
	if path == "" || path == "." {
		return y.root
	}
	n, ok := y.kvs[path]
	if ok {
		return n
	}
	return nil
}

func (y *YamlParser) All() map[string]*YamlNode {
	return y.kvs
}

func GenStdRegexp(pattern string) string {
	kkp := `(?U)\[(.*)\]`
	rr := regexp.MustCompile(kkp)
	for _, a := range rr.FindAllStringSubmatchIndex(pattern, -1) {
		ss := pattern[a[2]:a[3]]
		if _, err := strconv.Atoi(ss); err != nil {
			if ss == "*" {
				pattern = pattern[:a[0]] + `\[(.` + ss + `)\]` + pattern[a[1]:]
			} else {
				pattern = pattern[:a[0]] + `\[[` + ss + `]\]` + pattern[a[1]:]
			}
		}
	}
	return pattern
}

func (y *YamlParser) Match(pattern string) *YamlNode {
	if pattern == "" || pattern == "." {
		return y.root
	}
	result := &YamlNode{
		Name:     pattern,
		Path:     pattern,
		Ident:    y.Ident,
		Depth:    0,
		Line:     0,
		Value:    "",
		ArrNodes: nil,
		MapNodes: nil,
		prev:     nil,
		alias:    map[string]string{},
	}
	pattern = GenStdRegexp(pattern)
	r := regexp.MustCompile(pattern)
	for path, n := range y.kvs {
		if r.MatchString(path) {
			result.ArrNodes = append(result.ArrNodes, n)
		}
	}
	return result
}

func CountSpace(s string) int {
	var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}
	start := 0
	for ; start < len(s); start++ {
		c := s[start]
		if c >= utf8.RuneSelf {
			return start
		}
		if asciiSpace[c] == 0 {
			break
		}
	}
	return start
}

func isElem(s string) bool {
	return strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]")
}

func move(pointer *YamlNode, line []byte) *YamlNode {
	steps := (pointer.Depth*pointer.Ident - CountSpace(string(line))) / pointer.Ident
	if steps > 0 {
		for i := 0; i < steps; i++ {
			if pointer.prev == nil {
				break
			}
			if isElem(pointer.Name) {
				pointer.prev.ArrNodes = append(pointer.prev.ArrNodes, pointer)
			} else {
				pointer.prev.MapNodes = append(pointer.prev.MapNodes, pointer)
			}
			pointer = pointer.prev
		}
	}
	return pointer
}

func parseLine(line []byte) (ok bool, key, val string) {
	newStr := strings.TrimSpace(string(line))
	arr := strings.Split(newStr, ":")
	k := strings.TrimSpace(arr[0])
	var v string
	if len(arr) == 2 {
		v = strings.TrimSpace(arr[1])
		if len(v) > 0 && !strings.HasPrefix(v, "&") {
			return true, k, v
		}
	}
	if strings.HasPrefix(newStr, "-") && strings.HasSuffix(newStr, ":") {
		k += ":"
	}
	return false, k, v
}

func (y *YamlParser) add(n *YamlNode, path ...string) {
	var p = n.Path
	if len(path) > 0 {
		p = path[0]
	}
	y.kvs[p] = n
}

func (y *YamlParser) Parse(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	pointer := y.root
	pointer.prev = y.root
	n := 0
	var line []byte
	for s.Scan() {
		n++
		line = s.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		pointer = move(pointer, line)
		ok, k, v := parseLine(line)
		if k == "<<" {
			aliasNodeName := strings.Replace(v, "*", "&", 1)
			starNode, ok := y.alias[aliasNodeName]
			if !ok {
				continue
			}
			if len(starNode.ArrNodes) > 0 {
				s := len(pointer.ArrNodes)
				for i, an := range starNode.ArrNodes {
					aliasPath := fmt.Sprintf("%s.[%d].%s", pointer.Path, s+i, an.Name)
					y.add(an, aliasPath)
					pointer.alias[aliasPath] = an.Path
				}
			}
			if len(starNode.MapNodes) > 0 {
				for _, mn := range starNode.MapNodes {
					aliasPath := fmt.Sprintf("%s.%s", pointer.Path, mn.Name)
					y.add(mn, aliasPath)
					pointer.alias[aliasPath] = mn.Path
				}
			}
			pointer.addAliasNode(aliasNodeName, starNode)
		}
		var node *YamlNode
		if strings.HasPrefix(k, "-") {
			vv := strings.Split(k, " ")[1]
			vv = strings.TrimSpace(vv)
			kk := fmt.Sprintf("[%d]", len(pointer.ArrNodes))
			node = &YamlNode{
				Name:     kk,
				Path:     fmt.Sprintf("%s.%s", pointer.Path, kk),
				prev:     pointer,
				Ident:    pointer.Ident,
				Depth:    pointer.Depth + 1,
				Line:     n,
				Value:    "",
				MapNodes: nil,
				alias:    map[string]string{},
			}
			if len(v) > 0 {
				valNode := &YamlNode{
					Name:  vv,
					Path:  fmt.Sprintf("%s.%s", node.Path, vv),
					Ident: node.Ident,
					Depth: node.Depth,
					Line:  n,
					Value: v,
					prev:  node,
					alias: map[string]string{},
				}
				y.add(valNode)
				node.MapNodes = append(node.MapNodes, valNode)
				pointer = node
			} else {
				valNode := &YamlNode{
					Name:  vv,
					Path:  fmt.Sprintf("%s.%s", node.Path, vv),
					Ident: node.Ident,
					Depth: node.Depth,
					Line:  n,
					Value: "",
					prev:  node,
					alias: map[string]string{},
				}
				y.add(node)
				if strings.HasSuffix(vv, ":") {
					node.prev = pointer
					valNode.Name = strings.TrimRight(valNode.Name, ":")
					valNode.Path = strings.TrimRight(valNode.Path, ":")
					valNode.Depth += 1
					y.add(valNode)
					pointer = valNode
				} else {
					node.Value = vv
					pointer.ArrNodes = append(pointer.ArrNodes, node)
				}
			}
		}
		node = &YamlNode{
			Name:     k,
			Path:     fmt.Sprintf("%s.%s", pointer.Path, k),
			prev:     nil,
			Ident:    pointer.Ident,
			Depth:    pointer.Depth + 1,
			Line:     n,
			Value:    "",
			MapNodes: nil,
			alias:    map[string]string{},
		}
		node.prev = pointer
		if ok {
			if isElem(v) {
				arr := strings.Split(v[1:len(v)-1], ",")
				for i, a := range arr {
					aa := strings.TrimSpace(a)
					elemName := fmt.Sprintf("[%d]", i)
					elemPath := fmt.Sprintf("%s.%s", node.Path, elemName)
					elemNode := &YamlNode{
						Name:     elemName,
						Path:     elemPath,
						Ident:    y.Ident,
						Depth:    node.Depth + 1,
						Line:     n,
						Value:    aa,
						ArrNodes: nil,
						MapNodes: nil,
						prev:     node,
						alias:    map[string]string{},
					}
					y.add(elemNode)
					node.ArrNodes = append(node.ArrNodes, elemNode)
				}
				y.add(node)
				if isElem(pointer.Name) {
					pointer.ArrNodes = append(pointer.ArrNodes, node)
				} else {
					pointer.MapNodes = append(pointer.MapNodes, node)
				}
			} else {
				node.Value = v
				y.add(node)
				pointer.MapNodes = append(pointer.MapNodes, node)
			}
		} else {
			if strings.HasPrefix(v, "&") {
				node.Value = v
				y.alias[v] = node
			}
			pointer = node
			y.add(node)
		}
	}
	pointer = move(pointer, []byte{})
	return nil
}

func (y *YamlParser) Dump() {
	d, _ := json.MarshalIndent(y.root, "", "\t")
	println(string(d))
}
