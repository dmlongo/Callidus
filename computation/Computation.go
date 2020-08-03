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

func SubCSP_Computation(folderName string, domains map[string][]int, constraints []*Constraint, nodes []*Node,
	parallel bool, debugOption bool) bool {
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
	satisfiableChan := make(chan bool, len(nodes))
	satisfiable := true
	for _, node := range nodes {
		if parallel {
			go createAndSolveSubCSP(folderName, node, domains, constraints, wg, debugOption, satisfiableChan) //TODO: gestire possibile overhead
		} else {
			createAndSolveSubCSP(folderName, node, domains, constraints, wg, debugOption, satisfiableChan)
			satisfiable = <-satisfiableChan
			if !satisfiable {
				break
			}
		}
	}
	wg.Wait()
	cont := 0
	if parallel {
		cont = 0
		exit := false
		for !exit {
			select {
			case satisfiable = <-satisfiableChan:
				cont++
				if cont == len(nodes) || !satisfiable {
					exit = true
					break
				}

			}
		}
		close(satisfiableChan)
	}
	return satisfiable
}

func createAndSolveSubCSP(folderName string, node *Node, domains map[string][]int, constraints []*Constraint,
	wg *sync.WaitGroup, debugOption bool, satisfiableChan chan bool) {
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

	satisfiable := solve(fileName, debugOption)
	satisfiableChan <- satisfiable
	if satisfiable {
		AttachSingleNode(folderName, node, debugOption)
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

func solve(fileName string, debugOption bool) bool {
	defer func(debugOption bool) {
		if !debugOption {
			err := os.Remove(fileName)
			if err != nil {
				panic(err)
			}
		}
	}(debugOption)
	//Dalla wsl usare nacreWSL, da linux nativo usare nacre
	cmd := exec.Command("./libs/nacre", fileName, "-complete", "-sols", "-verb=3") //TODO: far funzionare nacre su windows
	out, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(out)
	var line string
	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	solFound := false
	parsedLine := make(chan string, 100)
	defer close(parsedLine)
	go func() {
		for {
			line, err = reader.ReadString('\n')
			if err == io.EOF && len(line) == 0 {
				break
			}
			if strings.HasPrefix(line, "v") {
				parsedLine <- parseLine(line)
				solFound = true
			}

		}
	}()
	outputFileName := strings.ReplaceAll(fileName, ".xml", "sol.txt")
	outfile, err := os.OpenFile(outputFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		panic(err)
	}
	for row := range parsedLine {
		_, err = outfile.WriteString(row + "\n")
		if err != nil {
			panic(err)
		}
	}
	if !solFound {
		return false
	}
	return true
}

func parseLine(line string) string {
	reg := regexp.MustCompile(".*<values>(.*) </values>.*")
	return reg.FindStringSubmatch(line)[1]
}

/*func doNacreMakeFile(){
	cmd := exec.Command("make")
	cmd.Dir = "/mnt/c/Users/simon/Desktop/Università/Tesi/Programmi/CSP_Project/libs/nacre_master/core"
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}*/

/*func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}*/
