package decomp

import (
	"os"
	"strconv"
	"sync"
)

func CreateAndSolveSubCSP(subCspFolder string, tablesFolder string, node *Node, domains map[string][]int, constraints []*Constraint,
	wg *sync.WaitGroup, satisfiableChan chan bool, subDebug bool, inMemory bool, parallel bool) bool {
	if parallel {
		defer wg.Done()
	}
	xmlFile := subCspFolder + strconv.Itoa(node.ID) + ".xml"
	tableFile := tablesFolder + strconv.Itoa(node.ID) + ".table"
	file, err := os.OpenFile(xmlFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		panic(err)
	}

	_, err = file.WriteString("<instance format=\"XCSP3\" type=\"CSP\">\n")
	if err != nil {
		panic(err)
	}
	writeVariables(file, node.Bag, domains)
	writeConstraints(file, node.Bag, constraints)
	_, err = file.WriteString("</instance>\n")
	if err != nil {
		panic(err)
	}
	err = file.Close()
	if err != nil {
		panic(err)
	}

	if parallel {
		satisfiable := solve(xmlFile, tableFile, satisfiableChan, node, subDebug, inMemory, parallel)
		satisfiableChan <- satisfiable
	} else {
		return solve(xmlFile, tableFile, nil, node, subDebug, inMemory, parallel)
	}
	return false
}

func writeVariables(file *os.File, variables []string, domains map[string][]int) {
	_, err := file.WriteString("<variables>\n")
	if err != nil {
		panic(err)
	}
	for _, variable := range variables {
		dom := domains[variable]
		values := "<var id=\"" + variable + "\"> "
		for _, i := range dom {
			values += strconv.Itoa(i) + " "
		}
		values += "</var>\n"
		_, err = file.WriteString(values)
		if err != nil {
			panic(err)
		}
	}
	_, err = file.WriteString("</variables>\n")

	if err != nil {
		panic(err)
	}
}

func writeConstraints(file *os.File, variables []string, constraints []*Constraint) {
	_, err := file.WriteString("<constraints>\n")
	if err != nil {
		panic(err)
	}
	for _, constraint := range constraints {
		if isConstraintOk(constraint.Scope, variables) {
			_, err = file.WriteString("<extension>\n")
			if err != nil {
				panic(err)
			}
			_, err = file.WriteString(getListVariable(constraint.Scope))
			if err != nil {
				panic(err)
			}
			_, err = file.WriteString(getPossibleValues(constraint))
			if err != nil {
				panic(err)
			}
			_, err = file.WriteString("</extension>\n")
			if err != nil {
				panic(err)
			}
		}
	}
	_, err = file.WriteString("</constraints>\n")
	if err != nil {
		panic(err)
	}
}

//check if the constraint is associated with a node variables
//TODO: we could use a map to speed up the check
func isConstraintOk(constraintVariables []string, variables []string) bool {
	ok := true
	for _, constraintVariable := range constraintVariables {
		isConstraintVariableOk := false
		for _, nodeVariable := range variables {
			if constraintVariable == nodeVariable {
				isConstraintVariableOk = true
				break
			}
		}
		if !isConstraintVariableOk {
			ok = false
			break
		}
	}
	return ok
}

func getListVariable(variables []string) string {
	listVariable := "<list> "
	for _, v := range variables {
		listVariable += v + " "
	}
	listVariable += "</list>\n"
	return listVariable
}
