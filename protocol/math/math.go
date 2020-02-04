package math

import (
	"context"
)

type Num struct {
	Val int32
}

type Counter interface {
	Inc(n1, n2 *Num) (n3 Num)
	Dec(n1, n2 Num) *Num
}

type Math interface {
	Counter
	Add(ctx context.Context, a, b int, n Num) (int, *Num)
	Sub(a, b int, n *Num) (cc int, nn Num)
	Sum(a int, n *Num, ctxs []context.Context, ns []int, is []interface{}, nns [][]int, nms []map[int]string, mns map[int][]string, ms map[string]interface{}, nm map[int]*Num, varargs ...interface{}) (x int, y map[int]map[string]interface{}, z map[int]*Num)
}
