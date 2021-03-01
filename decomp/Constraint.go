package decomp

// Constraint represents a constraint in a CSP
type Constraint struct {
	CType    bool //true if support, else conflict
	Scope    []string
	Relation [][]int
}

// AddVariable to this contraint scope
func (c *Constraint) AddVariable(v string) {
	c.Scope = append(c.Scope, v)
}

// AddTuple to this constraint relation
func (c *Constraint) AddTuple(tup []int) {
	c.Relation = append(c.Relation, tup)
}

// SetType of this constraint
func (c *Constraint) SetType(t bool) {
	c.CType = t
}
