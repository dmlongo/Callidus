package computation

import (
	. "../../CSP_Project/constraint"
	. "../../CSP_Project/hyperTree"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

//TODO: testare se è meglio una goroutine per ogni nodo oppure se è meglio suddividere il file in pezzi
func SubCSP_Computation(domains map[string][]int, constraints []*Constraint, nodes []*Node) {
	os.RemoveAll("subCSP")
	err := os.Mkdir("subCSP", 777)
	if err != nil {
		panic(err)
	}
	wg := &sync.WaitGroup{}
	wg.Add(len(nodes))
	for _, node := range nodes {
		go createAndSolveSubCSP(node, domains, constraints, wg)
	}
	wg.Wait()
}

func createAndSolveSubCSP(node *Node, domains map[string][]int, constraints []*Constraint, wg *sync.WaitGroup) {
	fileName := "subCSP/" + strconv.Itoa(node.Id) + ".xml"
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND, 777)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.WriteString("<instance format=\"XCSP3\" type=\"CSP\">\n")
	writeVariables(file, node.Variables, domains)
	writeConstraints(file, node.Variables, constraints)
	file.WriteString("</instance>\n")
	file.Close()

	solve(fileName)

	wg.Done()
}

func writeVariables(file *os.File, variables []string, domains map[string][]int) {
	file.WriteString("<variables>\n")
	for _, variable := range variables {
		dom := domains[variable]
		values := "<var id=\"" + variable + "\"> "
		for _, i := range dom {
			values += strconv.Itoa(i) + " "
		}
		values += "</var>\n"
		file.WriteString(values)
	}
	file.WriteString("</variables>\n")
}

func writeConstraints(file *os.File, variables []string, constraints []*Constraint) {
	file.WriteString("<constraints>\n")
	for _, constraint := range constraints {
		if isConstraintOk(constraint.Variables, variables) {
			file.WriteString("<extension>\n")
			file.WriteString(getListVariable(constraint.Variables))
			file.WriteString(getPossibleValues(constraint))
			file.WriteString("</extension>\n")
		}
	}
	file.WriteString("</constraints>\n")
}

//check if the constraint is associated with a node variables
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

func getPossibleValues(constraint *Constraint) string {
	possibleValues := "<supports> "
	if !constraint.CType {
		possibleValues = "<conflicts> "
	}
	for _, tup := range constraint.PossibleValues {
		possibleValues += "("
		for i := range tup {
			value := strconv.Itoa(tup[i])
			if i == len(tup)-1 {
				possibleValues += value
			} else {
				possibleValues += value + ","
			}
		}
		possibleValues += ")"
	}
	if constraint.CType {
		possibleValues += " </supports>\n"
	} else {
		possibleValues += " </conflicts>\n"
	}
	return possibleValues
}

func solve(fileName string) {
	outputFileName := strings.ReplaceAll(fileName, ".xml", "sol.txt")
	outputFile, err := os.OpenFile(outputFileName, os.O_CREATE|os.O_APPEND, 777)
	if err != nil {
		panic(err)
	}
	out, err := exec.Command("java", "-cp", "libs/AbsCon.jar", "AbsCon", fileName, "-s=all").Output()
	if err != nil {
		panic(err)
	}
	outputFile.Write(out)
}
