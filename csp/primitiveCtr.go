package csp

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
	c.strVars = append(c.strVars, strings.Split(c.Vars, " ")...)
	return c.strVars
}

// ToXCSP converts this constraint in the XCSP format
func (c *PrimitiveCtr) ToXCSP() []string {
	return []string{"<intension> " + c.Function + " </intension>"}
}
