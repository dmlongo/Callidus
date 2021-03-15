package ctr

import (
	"os"
	"regexp"
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
	varList := makeVarList(vars)

	var sb strings.Builder
	sb.WriteString("<instantiation>\n")
	sb.WriteString("\t<list> ")
	for _, v := range varList {
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

func makeVarList(sortedVars []string) []string {
	var list []string
	var arrayID = regexp.MustCompile(`^((\w)+?)((L\d+J)+)$`)
	var indices = regexp.MustCompile(`L\d+J`)

	var sb strings.Builder
	for _, v := range sortedVars {
		if arrayID.MatchString(v) {
			tks := arrayID.FindStringSubmatch(v)
			name := tks[1]
			if list == nil || !strings.HasPrefix(list[len(list)-1], name+"[") {
				sb.WriteString(name)
				for range indices.FindAllString(v, -1) {
					sb.WriteString("[]")
				}
				list = append(list, sb.String())
				sb.Reset()
			}
		} else {
			list = append(list, v)
		}
	}

	return list
}
