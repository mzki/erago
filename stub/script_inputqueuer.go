package stub

import "strconv"

// implements script.InputQueuer
type scriptInputQueuer struct {
	xs []string
}

func NewScriptInputQueuer() *scriptInputQueuer { return &scriptInputQueuer{make([]string, 0)} }

func (q *scriptInputQueuer) Append(x string) (n int) {
	q.xs = append(q.xs, x)
	return len(q.xs)
}

func (q *scriptInputQueuer) Prepend(x string) (n int) {
	q.xs = append([]string{x}, q.xs...)
	return len(q.xs)
}

func (q *scriptInputQueuer) Clear()    { q.xs = []string{} }
func (q *scriptInputQueuer) Size() int { return len(q.xs) }

func (q *scriptInputQueuer) TryDeque() (string, bool) {
	if len(q.xs) > 0 {
		x := q.xs[0]
		q.xs = q.xs[1:]
		return x, true
	} else {
		return "", false
	}
}

func (q *scriptInputQueuer) Deque() string {
	ret, _ := q.TryDeque()
	return ret
}

func (q *scriptInputQueuer) TryDequeInt() (int, bool) {
	cmd := q.Deque()
	ret, err := strconv.Atoi(cmd)
	if err == nil {
		return ret, true
	} else {
		return 0, false
	}
}

func (q *scriptInputQueuer) DequeInt() int {
	ret, _ := q.TryDequeInt()
	return ret
}
