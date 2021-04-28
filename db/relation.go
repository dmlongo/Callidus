package db

import (
	"fmt"
)

// Relation represent a set of tuples
type Relation interface {
	Attributes() []string
	Position(attr string) (int, bool)
	AddTuple(vals []int) (Tuple, bool)
	RemoveTuples(idx []int) (bool, error)
	Tuples() []Tuple
	Empty() bool
}

// Tuple represent a row in a relation
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
	// TODO no check if the tuple is already here
	t.tuples = append(t.tuples, vals)
	return vals, true
}

func (t *table) RemoveTuples(idx []int) (bool, error) {
	if len(idx) == 0 {
		return false, nil
	}
	newSize := len(t.tuples) - len(idx)
	if newSize < 0 {
		return false, fmt.Errorf("new size %v < 0", newSize)
	}
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
	return true, nil
}

func (t *table) Tuples() []Tuple {
	return t.tuples
}
