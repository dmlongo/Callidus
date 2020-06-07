package constraint

type Constraint struct {
	CType          bool //true if support, else conflict
	Variables      []string
	PossibleValues [][]int
}

func (c *Constraint) AddVariable(v string) {
	c.Variables = append(c.Variables, v)
}

func (c *Constraint) AddPossibleValue(tup []int) {
	c.PossibleValues = append(c.PossibleValues, tup)
}

func (c *Constraint) SetType(t bool) {
	c.CType = t
}
