package ctr

import "strings"

// PrimitiveCtr represents a primitive constraint in XCSP
type PrimitiveCtr struct {
	CName    string
	Vars     string
	strVars  []string
	Function string
}

// Name of this constraint
func (c *PrimitiveCtr) Name() string {
	return c.CName
}

// Variables of this constraint
func (c *PrimitiveCtr) Variables() []string {
	if c.strVars != nil {
		return c.strVars
	}
	for _, v := range strings.Split(c.Vars, " ") {
		c.strVars = append(c.strVars, v)
	}
	return c.strVars
}

// ToXCSP converts this constraint in the XCSP format
func (c *PrimitiveCtr) ToXCSP() []string {
	return []string{"<intension> " + c.Function + " </intension>"}
}
