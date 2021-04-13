package decomp

// Relation represent a set of tuples
type Relation interface {
	Attributes() []string
	Position(attr string) (int, bool)
	AddTuple(vals []int) (Tuple, bool)
	RemoveTuples(idx []int) bool
	Tuples() []Tuple
	Empty() bool
}

type Tuple []int

type table struct {
	attrs   []string
	attrPos map[string]int
	tuples  []Tuple
}

func NewRelation(attrs []string) Relation {
	if len(attrs) <= 0 {
		return nil
	}
	attrPos := make(map[string]int)
	for i, v := range attrs {
		attrPos[v] = i
	}
	return &table{attrs, attrPos, make([]Tuple, 0)}
}

func InitializedRelation(attrs []string, rel []Tuple) Relation {
	if len(attrs) <= 0 {
		return nil
	}
	attrPos := make(map[string]int)
	for i, v := range attrs {
		attrPos[v] = i
	}
	return &table{attrs, attrPos, rel}
}

func (t *table) Empty() bool {
	return len(t.tuples) == 0
}

func (t *table) Attributes() []string {
	return t.attrs
}

func (t *table) Position(attr string) (pos int, ok bool) {
	pos, ok = t.attrPos[attr]
	return
}

func (t *table) AddTuple(vals []int) (Tuple, bool) {
	if len(t.attrs) != len(vals) {
		return nil, false
	}
	// TODO check domains?
	t.tuples = append(t.tuples, vals)
	return vals, true
}

func (t *table) RemoveTuples(idx []int) bool {
	if len(idx) == 0 {
		return false
	}
	newSize := len(t.tuples) - len(idx)
	newTuples := make([]Tuple, 0, newSize)
	if newSize > 0 {
		i := 0
		for _, j := range idx {
			newTuples = append(newTuples, t.tuples[i:j]...)
			i = j + 1
		}
		newTuples = append(newTuples, t.tuples[i:]...)
	}
	t.tuples = newTuples
	return true
}

func (t *table) Tuples() []Tuple {
	return t.tuples
}

type Filter func(t Tuple) bool

func Semijoin(l Relation, r Relation) (Relation, bool) {
	joinIdx := commonAttrs(l, r)

	var tupToDel []int
	for i, leftTup := range l.Tuples() {
		delete := true
		for _, rightTup := range r.Tuples() {
			if match(leftTup, rightTup, joinIdx) {
				delete = false
				break
			}
		}
		if delete {
			tupToDel = append(tupToDel, i)
		}
	}

	res := l.RemoveTuples(tupToDel)
	return l, res
}

func Join(l Relation, r Relation) Relation {
	newAttrs := joinedAttrs(l, r)
	joinIdx := commonAttrs(l, r)
	newRel := NewRelation(newAttrs)
	for _, lTup := range l.Tuples() {
		for _, rTup := range r.Tuples() {
			if match(lTup, rTup, joinIdx) {
				newTup := joinedTuple(newAttrs, lTup, rTup, r.Position)
				newRel.AddTuple(newTup)
			}
		}
	}
	return newRel
}

func Select(r Relation, f Filter) (Relation, bool) {
	var tupToDel []int
	for i, tup := range r.Tuples() {
		if f(tup) {
			tupToDel = append(tupToDel, i)
		}
	}
	res := r.RemoveTuples(tupToDel)
	return r, res
}

func commonAttrs(left Relation, right Relation) [][]int {
	var out [][]int
	rev := len(right.Attributes()) < len(left.Attributes())
	if rev {
		left, right = right, left
	}
	for iLeft, varLeft := range left.Attributes() {
		if iRight, found := right.Position(varLeft); found {
			if rev {
				out = append(out, []int{iRight, iLeft})
			} else {
				out = append(out, []int{iLeft, iRight})
			}
		}
	}
	return out
}

func match(left Tuple, right Tuple, joinIndex [][]int) bool {
	for _, z := range joinIndex {
		if left[z[0]] != right[z[1]] {
			return false
		}
	}
	return true
}

func joinedAttrs(l Relation, r Relation) []string {
	var res []string
	res = append(res, l.Attributes()...)
	for _, v := range r.Attributes() {
		if _, found := l.Position(v); !found {
			res = append(res, v)
		}
	}
	return res
}

func joinedTuple(attrs []string, lTup Tuple, rTup Tuple, rPos func(string) (int, bool)) Tuple {
	res := make([]int, 0, len(attrs))
	res = append(res, lTup...)
	for _, v := range attrs[len(lTup):] {
		i, _ := rPos(v)
		res = append(res, rTup[i])
	}
	return res
}
