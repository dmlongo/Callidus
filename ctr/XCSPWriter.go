package ctr

import (
	"os"
	"sort"
	"strconv"
	"strings"
)

// TODO maybe instead of writing to file, I could just return strings (for in-memory reasons)

// CreateXCSPInstance from given constraints
func CreateXCSPInstance(constraints []Constraint, variables map[string]string, outFile string) {
	file, err := os.OpenFile(outFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		panic(err)
	}
	_, err = file.WriteString("<instance format=\"XCSP3\" type=\"CSP\">\n")
	if err != nil {
		panic(err)
	}
	writeVariables(file, variables)
	writeConstraints(file, constraints)
	_, err = file.WriteString("</instance>\n")
	if err != nil {
		panic(err)
	}
	err = file.Close()
	if err != nil {
		panic(err)
	}
}

func writeVariables(file *os.File, variables map[string]string) {
	_, err := file.WriteString("\t<variables>\n")
	if err != nil {
		panic(err)
	}
	var vars []string
	for v := range variables {
		vars = append(vars, v)
	}
	sort.Strings(vars)
	for _, v := range vars {
		values := "\t\t<var id=\"" + v + "\"> "
		values += variables[v]
		values += " </var>\n"
		_, err = file.WriteString(values)
		if err != nil {
			panic(err)
		}
	}
	_, err = file.WriteString("\t</variables>\n")
	if err != nil {
		panic(err)
	}
}

func writeConstraints(file *os.File, constraints []Constraint) {
	_, err := file.WriteString("\t<constraints>\n")
	if err != nil {
		panic(err)
	}
	for _, c := range constraints {
		for _, line := range c.ToXCSP() {
			_, err := file.WriteString("\t\t" + line + "\n")
			if err != nil {
				panic(err)
			}
		}
	}
	_, err = file.WriteString("\t</constraints>\n")
	if err != nil {
		panic(err)
	}
}

// WriteSolution in XCSP format
func WriteSolution(sol Solution) string {
	vars := sol.sortVars()
	// TODO convert many-vars to array

	var sb strings.Builder
	sb.WriteString("<instantiation>\n")
	sb.WriteString("\t<list> ")
	for _, v := range vars {
		sb.WriteString(v)
		sb.WriteString(" ")
	}
	sb.WriteString("</list>\n")
	sb.WriteString("\t<values> ")
	for _, v := range vars {
		val := sol[v]
		sb.WriteString(strconv.Itoa(val))
		sb.WriteString(" ")
	}
	sb.WriteString("</values>\n")
	sb.WriteString("</instantiation>\n")
	return sb.String()
}
