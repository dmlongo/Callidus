package computation

import (
	. "../../Callidus/constraint"
	. "../../Callidus/hyperTree"
	"bufio"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

//TODO: testare se è meglio una goroutine per ogni nodo oppure se è meglio suddividere il file in pezzi
func SubCSP_Computation(folderName string, domains map[string][]int, constraints []*Constraint, nodes []*Node, parallel bool) {
	//doNacreMakeFile() //se la scelta del solver è nacre
	err := os.RemoveAll(folderName)
	if err != nil {
		panic(err)
	}
	err = os.Mkdir(folderName, 0777)
	if err != nil {
		panic(err)
	}
	wg := &sync.WaitGroup{}
	wg.Add(len(nodes))
	for _, node := range nodes {
		if parallel {
			go createAndSolveSubCSP(folderName, node, domains, constraints, wg) //TODO: gestire possibile overhead
		} else {
			createAndSolveSubCSP(folderName, node, domains, constraints, wg)
		}
	}
	wg.Wait()
}

func createAndSolveSubCSP(folderName string, node *Node, domains map[string][]int, constraints []*Constraint, wg *sync.WaitGroup) {
	defer wg.Done()
	fileName := folderName + strconv.Itoa(node.Id) + ".xml"
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

	solve(fileName)
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

func solve(fileName string) {
	//Dalla wsl usare nacreWSL, da linux nativo usare nacre
	cmd := exec.Command("./libs/nacre", fileName, "-complete", "-sols", "-verb=3") //TODO: far funzionare nacre su windows
	out, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(out)
	var line string
	result := make(map[string][]string)
	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	for {
		line, err = reader.ReadString('\n')
		if err == io.EOF && len(line) == 0 {
			break
		}
		if strings.HasPrefix(line, "v") {
			parseLine(line, result)
		}
	}
	outputFileName := strings.ReplaceAll(fileName, ".xml", "sol.txt")
	outfile, err := os.OpenFile(outputFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		panic(err)
	}
	for key, values := range result {
		_, err = outfile.WriteString(key + " ->")
		if err != nil {
			panic(err)
		}
		for _, v := range values {
			_, err = outfile.WriteString(" " + v)
			if err != nil {
				panic(err)
			}
		}
		_, err = outfile.WriteString("\n")
		if err != nil {
			panic(err)
		}
	}
}

func parseLine(line string, result map[string][]string) {
	reg := regexp.MustCompile(".*<list> (.*) </list> <values>(.*) </values>.*")
	keys := strings.Split(reg.FindStringSubmatch(line)[1], " ")
	values := strings.Split(reg.FindStringSubmatch(line)[2], " ")
	for i, k := range keys {
		result[k] = append(result[k], values[i])
	}
}

/*func doNacreMakeFile(){
	cmd := exec.Command("make")
	cmd.Dir = "/mnt/c/Users/simon/Desktop/Università/Tesi/Programmi/CSP_Project/libs/nacre_master/core"
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}*/
