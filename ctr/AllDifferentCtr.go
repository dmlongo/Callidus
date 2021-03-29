package ctr

import "strings"

// AllDifferentCtr represents an allDifferent constraint in XCSP
type AllDifferentCtr struct {
	CName   string
	Vars    string
	strVars []string
}

// Name of this constraint
func (c *AllDifferentCtr) Name() string {
	return c.CName
}

// Variables of this constraint
func (c *AllDifferentCtr) Variables() []string {
	if c.strVars != nil {
		return c.strVars
	}
	c.strVars = append(c.strVars, strings.Split(c.Vars, " ")...)
	return c.strVars
}

// ToXCSP converts this constraint in the XCSP format
func (c *AllDifferentCtr) ToXCSP() []string {
	return []string{"<allDifferent> " + c.Vars + " </allDifferent>"}
}
