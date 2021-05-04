package csp

import "strings"

// allDifferentCtr represents an allDifferent constraint in XCSP
type allDifferentCtr struct {
	CName   string
	Vars    string
	strVars []string
}

// Name of this constraint
func (c *allDifferentCtr) Name() string {
	return c.CName
}

// Variables of this constraint
func (c *allDifferentCtr) Variables() []string {
	if c.strVars != nil {
		return c.strVars
	}
	c.strVars = append(c.strVars, strings.Split(c.Vars, " ")...)
	return c.strVars
}

// ToXCSP converts this constraint in the XCSP format
func (c *allDifferentCtr) ToXCSP() []string {
	return []string{"<allDifferent> " + c.Vars + " </allDifferent>"}
}
