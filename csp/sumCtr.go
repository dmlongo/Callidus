package csp

import "strings"

// sumCtr represents a sum constraint in XCSP
type sumCtr struct {
	CName     string
	Vars      string
	strVars   []string
	Coeffs    string
	Condition string
}

// Name of this constraint
func (c *sumCtr) Name() string {
	return c.CName
}

// Variables of this constraint
func (c *sumCtr) Variables() []string {
	if c.strVars != nil {
		return c.strVars
	}
	c.strVars = append(c.strVars, strings.Split(c.Vars, " ")...)
	return c.strVars
}

// ToXCSP converts this constraint in the XCSP format
func (c *sumCtr) ToXCSP() []string {
	out := make([]string, 0, 5)
	out = append(out, "<sum>")
	out = append(out, "\t<list> "+c.Vars+" </list>")
	out = append(out, "\t<coeffs> "+c.Coeffs+" </coeffs>")
	out = append(out, "\t<condition> "+c.Condition+" </condition>")
	out = append(out, "</sum>")
	return out
}
