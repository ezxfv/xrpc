package config

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v2"
)

type ValueType int

const (
	IntVal ValueType = iota
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

	prev  *YamlNode
	alias map[string]string
}

func (n *YamlNode) Marshal(v interface{}) (err error) {
	w := bytes.NewBuffer([]byte(nil))
	n.Dump(w)
	println(w.String())
	return yaml.Unmarshal(w.Bytes(), v)
}

func (n *YamlNode) value(t ValueType) interface{} {
	switch t {
	case StringVal:
		return n.Value
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
	return &YamlNode{}
}

func (n *YamlNode) String() string {
	if len(n.Value) == 0 {
		return ""
	}
	return n.value(StringVal).(string)
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

func (n *YamlNode) IntArray() []int {
	if len(n.ArrNodes) == 0 {
		return []int{}
	}
	is := n.value(ArrayVal).([]*YamlNode)
	var res []int
	for _, s := range is {
		v, err := strconv.Atoi(s.Value)
		if err != nil {
			continue
		}
		res = append(res, v)
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
		v, err := strconv.ParseFloat(s.Value, 10)
		if err != nil {
			continue
		}
		res = append(res, v)
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
		res = append(res, s.Value)
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

func (n *YamlNode) Dump(w io.Writer, start ...int) {
	s := n.Depth
	if len(start) > 0 {
		s = start[0]
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
		return
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
			} else {
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
	}
}

type YamlParser struct {
	Ident int
	alias map[string]*YamlNode
	root  *YamlNode
}

func (y *YamlParser) Get(path string) *YamlNode {
	if path == "" || path == "." {
		return y.root
	}
	for _, c := range y.root.MapNodes {
		if c.Path == path {
			return c
		}
		if strings.HasPrefix(path, c.Path) {
			return c.Get(path)
		}
	}
	return &YamlNode{}
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
			starNode, ok := y.alias[strings.Replace(v, "*", "&", 1)]
			if !ok {
				continue
			}
			if len(starNode.ArrNodes) > 0 {
				s := len(pointer.ArrNodes)
				pointer.ArrNodes = append(pointer.ArrNodes, starNode.ArrNodes...)
				for i := 0; i < len(starNode.ArrNodes); i++ {
					pointer.alias[fmt.Sprintf("%s.[%d].%s", pointer.Path, s+i, starNode.ArrNodes[i].Name)] = starNode.ArrNodes[i].Path
				}
			}
			if len(starNode.MapNodes) > 0 {
				pointer.MapNodes = append(pointer.MapNodes, starNode.MapNodes...)
				for _, mn := range starNode.MapNodes {
					pointer.alias[fmt.Sprintf("%s.%s", pointer.Path, mn.Name)] = mn.Path
				}
			}
			continue
		}
		var node *YamlNode
		if strings.HasPrefix(k, "-") {
			vv := strings.Split(k, " ")[1]
			vv = strings.TrimSpace(vv)
			kk := fmt.Sprintf("[%d]", len(pointer.ArrNodes))
			node = &YamlNode{
				Name:     kk,
				Path:     fmt.Sprintf("%s.%s", pointer.Path, kk),
				prev:     nil,
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
				node.MapNodes = append(node.MapNodes, valNode)
				pointer.ArrNodes = append(pointer.ArrNodes, node)
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
				if strings.HasSuffix(vv, ":") {
					node.prev = pointer
					valNode.Name = strings.TrimRight(valNode.Name, ":")
					valNode.Path = strings.TrimRight(valNode.Path, ":")
					valNode.Depth += 1
					pointer = valNode
				} else {
					node.Value = vv
					pointer.ArrNodes = append(pointer.ArrNodes, node)
				}
			}
			continue
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
					node.ArrNodes = append(node.ArrNodes, elemNode)
				}
				pointer.MapNodes = append(pointer.MapNodes, node)
			} else {
				node.Value = v
				pointer.MapNodes = append(pointer.MapNodes, node)
			}
		} else {
			pointer = node
			if strings.HasPrefix(v, "&") {
				y.alias[v] = pointer
			}
		}
	}
	pointer = move(pointer, []byte{})
	return nil
}

func (y *YamlParser) Dump() {
	d, _ := json.MarshalIndent(y.root, "", "\t")
	println(string(d))
}
