package db

type Condition func(t Tuple) bool

func Semijoin(l Relation, r Relation) (Relation, bool) {
	joinIdx := commonAttrs(l, r)
	if len(joinIdx) == 0 {
		return l, false
	}

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

	res, err := l.RemoveTuples(tupToDel)
	if err != nil {
		panic(err)
	}
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

func Select(r Relation, c Condition) (Relation, bool) {
	var tupToDel []int
	for i, tup := range r.Tuples() {
		if !c(tup) {
			tupToDel = append(tupToDel, i)
		}
	}
	res, err := r.RemoveTuples(tupToDel)
	if err != nil {
		panic(err)
	}
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
