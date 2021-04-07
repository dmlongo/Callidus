package ctr

import "strings"

// ElementCtr represents an element constraint in XCSP
type ElementCtr struct {
	CName      string
	Vars       string
	strVars    []string
	List       string
	StartIndex string
	Index      string
	Rank       string
	Condition  string
}

// Name of this constraint
func (c *ElementCtr) Name() string {
	return c.CName
}

// Variables of this constraint
func (c *ElementCtr) Variables() []string {
	if c.strVars != nil {
		return c.strVars
	}
	c.strVars = append(c.strVars, strings.Split(c.Vars, " ")...)
	return c.strVars
}

// ToXCSP converts this constraint in the XCSP format
func (c *ElementCtr) ToXCSP() []string {
	out := make([]string, 0, 5)
	out = append(out, "<element>")
	out = append(out, "\t<list startIndex=\""+c.StartIndex+"\"> "+c.List+" </list>")
	out = append(out, "\t<index rank=\""+c.Rank+"\"> "+c.Index+" </index>")
	out = append(out, "\t<value> "+c.Condition+" </value>")
	out = append(out, "</element>")
	return out
}
