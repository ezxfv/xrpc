package math

type Num struct {
	Val int32
	S   Step
}

type Step struct {
	S int32
}

type Counter interface {
	Inc(n *Num) (int32, *Num)
	Dec(n Num) *Num
}

type Math interface {
	Counter
	Add(a, b int) int
	Calc(ints ...int) (int, float64)
	//Sum(a int, n *Num, ctxs []context.Context, ns []int, is []interface{}, nns [][]int, nms []map[int]string, mns map[int][]string, ms map[string]interface{}, nm map[int]*Num, varargs ...interface{}) (x int, y map[int]map[string]interface{}, z map[int]*Num)
}
