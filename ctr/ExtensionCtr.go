package ctr

import "strings"

// ExtensionCtr represents an extensional constraint in XCSP
type ExtensionCtr struct {
	CName   string
	Vars    string
	strVars []string
	CType   string
	Tuples  string
}

// Name of this constraint
func (c *ExtensionCtr) Name() string {
	return c.CName
}

// Variables of this constraint
func (c *ExtensionCtr) Variables() []string {
	if c.strVars != nil { // TODO is this nil or a zero value?
		return c.strVars
	}
	c.strVars = make([]string, 0)
	for _, v := range strings.Split(c.Vars, " ") {
		c.strVars = append(c.strVars, v)
	}
	return c.strVars
}

// ToXCSP converts this constraint in the XCSP format
func (c *ExtensionCtr) ToXCSP() []string {
	out := make([]string, 0, 4)
	out = append(out, "<extension>")
	out = append(out, "\t<list> "+c.Vars+" </list>")
	out = append(out, "\t<"+c.CType+"> "+c.Tuples+" </"+c.CType+">")
	out = append(out, "</extension>")
	return out
}

// AddVariable to this contraint scope
/*func (c *ExtensionCtr) AddVariable(v string) {
	c.Vars = append(c.Vars, v)
}*/

// AddTuple to this constraint relation
/*func (c *ExtensionCtr) AddTuple(tup []int) {
	c.Tuples = append(c.Tuples, tup)
}*/

// SetType of this constraint
/*func (c *ExtensionCtr) SetType(t bool) {
	c.CType = t
}*/
