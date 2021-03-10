package ctr

import (
	"os"
)

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
	for v, dom := range variables {
		values := "\t\t<var id=\"" + v + "\"> "
		values += dom
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
