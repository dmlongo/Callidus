package computation

import (
	. "../../CSP_Project/constraint"
	. "../../CSP_Project/hyperTree"
	"strings"

	"os"
	"os/exec"
	"strconv"
	"sync"
)

//TODO: testare se è meglio una goroutine per ogni nodo oppure se è meglio suddividere il file in pezzi
func SubCSP_Computation(domains map[string][]int, constraints []*Constraint, nodes []*Node, inMemory bool,
	solver string, parallel bool) []string {
	//doNacreMakeFile() //se la scelta del solver è nacre
	solutions := make([]string, len(nodes))
	err := os.RemoveAll("subCSP")
	if err != nil {
		panic(err)
	}
	err = os.Mkdir("subCSP", 0777)
	if err != nil {
		panic(err)
	}
	wg := &sync.WaitGroup{}
	wg.Add(len(nodes))
	for i, node := range nodes {
		c := make(chan string, 1)
		if parallel {
			go createAndSolveSubCSP(node, domains, constraints, wg, c, inMemory, solver) //TODO: gestire possibile overhead
		} else {
			createAndSolveSubCSP(node, domains, constraints, wg, c, inMemory, solver)
		}
		if inMemory {
			sol := <-c
			solutions[i] = sol
		}
	}
	wg.Wait()
	return solutions
}

func createAndSolveSubCSP(node *Node, domains map[string][]int, constraints []*Constraint, wg *sync.WaitGroup,
	c chan string, inMemory bool, solver string) {
	defer wg.Done()
	fileName := "subCSP/" + strconv.Itoa(node.Id) + ".xml"
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		panic(err)
	}

	_, err = file.WriteString("<instance format=\"XCSP3\" type=\"CSP\">\n")
	if err != nil {
		panic(err)
	}
	writeVariables(file, node.Variables, domains)
	writeConstraints(file, node.Variables, constraints)
	_, err = file.WriteString("</instance>\n")
	if err != nil {
		panic(err)
	}
	err = file.Close()
	if err != nil {
		panic(err)
	}

	result := solve(fileName, inMemory, solver)

	if inMemory {
		c <- result
	}
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
		if isConstraintOk(constraint.Variables, variables) {
			_, err = file.WriteString("<extension>\n")
			if err != nil {
				panic(err)
			}
			_, err = file.WriteString(getListVariable(constraint.Variables))
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

func solve(fileName string, inMemory bool, solver string) string {
	//Dalla wsl usare nacreWSL, da linux nativo usare nacre
	var cmd *exec.Cmd
	if solver == "Nacre" {
		cmd = exec.Command("./libs/nacreWSL", fileName, "-complete", "-sols", "-verb=3") //TODO: far funzionare nacre su windows
	} else if solver == "AbsCon" {
		cmd = exec.Command("java", "-cp", "./libs/AbsCon.jar", "AbsCon", fileName, "-s=all")
	} else {
		panic("solver not found")
	}
	if inMemory {
		buffer, _ := cmd.Output()
		return string(buffer)
	} else {
		outputFileName := strings.ReplaceAll(fileName, ".xml", "sol.txt")
		outfile, err := os.Create(outputFileName)
		if err != nil {
			panic(err)
		}
		cmd.Stdout = outfile
		err = cmd.Run()
		err = outfile.Close()
		if err != nil {
			panic(err)
		}
		return ""
	}
}

/*func doNacreMakeFile(){
	cmd := exec.Command("make")
	cmd.Dir = "/mnt/c/Users/simon/Desktop/Università/Tesi/Programmi/CSP_Project/libs/nacre_master/core"
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}*/
