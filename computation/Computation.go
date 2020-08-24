package computation

import (
	. "../../Callidus/constraint"
	. "../../Callidus/hyperTree"
	. "../../Callidus/pre-processing"
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

func SubCSP_Computation(domains map[string][]int, constraints []*Constraint, nodes []*Node) bool {
	subCspFolder := "subCSP-" + SystemSettings.FolderName
	tablesFolder := "tables-" + SystemSettings.FolderName
	err := os.RemoveAll(subCspFolder)
	if err != nil {
		panic(err)
	}
	err = os.RemoveAll(tablesFolder)
	if err != nil {
		panic(err)
	}
	err = os.Mkdir(subCspFolder, 0777)
	if err != nil {
		panic(err)
	}
	err = os.Mkdir(tablesFolder, 0777)
	if err != nil {
		panic(err)
	}
	satisfiable := true
	if SystemSettings.ParallelSC {
		subCspSemaphore := &sync.WaitGroup{}
		subCspSemaphore.Add(len(nodes))
		checkSatisfiableSemaphore := &sync.WaitGroup{}
		checkSatisfiableSemaphore.Add(1)
		satisfiableChan := make(chan bool, len(nodes))
		for _, node := range nodes {
			go createAndSolveSubCSP(subCspFolder, tablesFolder, node, domains, constraints, subCspSemaphore, satisfiableChan)
		}
		go func() {
			defer checkSatisfiableSemaphore.Done()
			for satisfiable = range satisfiableChan {
				if !satisfiable {
					break
				}
			}
		}()
		subCspSemaphore.Wait()
		close(satisfiableChan)
		checkSatisfiableSemaphore.Wait()
	} else {
		for _, node := range nodes {
			satisfiable = createAndSolveSubCSP(subCspFolder, tablesFolder, node, domains, constraints, nil, nil)
			if !satisfiable {
				break
			}
		}
	}
	return satisfiable
}

func createAndSolveSubCSP(subCspFolder string, tablesFolder string, node *Node, domains map[string][]int, constraints []*Constraint,
	wg *sync.WaitGroup, satisfiableChan chan bool) bool {
	if SystemSettings.ParallelSC {
		defer wg.Done()
	}
	xmlFile := subCspFolder + strconv.Itoa(node.Id) + ".xml"
	tableFile := tablesFolder + strconv.Itoa(node.Id) + ".table"
	file, err := os.OpenFile(xmlFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
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
	fileStats, err := file.Stat()
	if err != nil {
		panic(err)
	}
	fileSizeKB := fileStats.Size() / 1024
	err = file.Close()
	if err != nil {
		panic(err)
	}

	if SystemSettings.ParallelSC {
		satisfiable := solve(xmlFile, tableFile, satisfiableChan, node, fileSizeKB)
		satisfiableChan <- satisfiable
	} else {
		return solve(xmlFile, tableFile, nil, node, fileSizeKB)
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

func solve(xmlFile string, tableFile string, satisfiableChan chan bool, node *Node, fileSizeKB int64) bool {
	defer func(debugOption bool) {
		if !debugOption {
			err := os.Remove(xmlFile)
			if err != nil {
				panic(err)
			}
		}
	}(SystemSettings.Debug)
	fmt.Println(fileSizeKB)
	if fileSizeKB > 1000 {
		cmd := exec.Command("./Callidus", xmlFile, "-i")
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

		solFound := true
		var outputTable *os.File = nil
		if SystemSettings.InMemory {
			node.PossibleValues = make([][]int, 0)
		} else {
			outputTable, err = os.OpenFile(tableFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
		}
		if err != nil {
			panic(err)
		}

		if SystemSettings.ParallelSC {
			exit := false
			for !exit {
				select {
				case check := <-satisfiableChan:
					if !check {
						err = cmd.Process.Kill()
						if err != nil {
							panic(err)
						}
						exit = true
						break
					}
				default:
					line, err = reader.ReadString('\n')
					if err == io.EOF && len(line) == 0 {
						exit = true
						break
					}
					if strings.HasPrefix(line, "N") {
						solFound = false
					} else if strings.HasPrefix(line, "Sol") {
						for i := 0; i < len(node.Variables); i++ {
							line, err = reader.ReadString('\n')
							outputTable.WriteString(strings.Split(line, " ")[len(line)-1])
							fmt.Println(strings.Split(line, " ")[len(line)-1])
						}
						outputTable.WriteString("\n")
						fmt.Println(" ")
					}
				}
			}

		} else {
			for {
				line, err = reader.ReadString('\n')
				if err == io.EOF && len(line) == 0 {
					break
				}
				if strings.HasPrefix(line, "N") {
					solFound = false
				} else if strings.HasPrefix(line, "Sol") {
					temp := make([]int, len(node.Variables))
					for i := 0; i < len(node.Variables); i++ {
						line, err = reader.ReadString('\n')
						if err != nil {
							panic(err)
						}
						val, err := strconv.Atoi(strings.Split(line, " ")[len(line)-1])
						if err != nil {
							panic(err)
						}
						temp[i] = val
					}
					node.PossibleValues = append(node.PossibleValues, temp)
				}
			}
		}
		return solFound
	} else {
		cmd := exec.Command("./libs/nacreWSL", xmlFile, "-complete", "-sols", "-verb=3")
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
		var outputTable *os.File = nil
		if SystemSettings.InMemory {
			node.PossibleValues = make([][]int, 0)
		} else {
			outputTable, err = os.OpenFile(tableFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
		}
		if err != nil {
			panic(err)
		}
		if SystemSettings.ParallelSC {
			exit := false
			for !exit {
				select {
				case check := <-satisfiableChan:
					if !check {
						err = cmd.Process.Kill()
						if err != nil {
							panic(err)
						}
						exit = true
						break
					}
				default:
					line, err = reader.ReadString('\n')
					if err == io.EOF && len(line) == 0 {
						exit = true
						break
					}
					if strings.HasPrefix(line, "v") {
						reg := regexp.MustCompile(".*<values>(.*) </values>.*")
						if SystemSettings.InMemory {
							temp := make([]int, len(node.Variables))
							for i, value := range strings.Split(reg.FindStringSubmatch(line)[1], " ") {
								v, err := strconv.Atoi(value)
								if err != nil {
									panic(err)
								}
								temp[i] = v
							}
							node.PossibleValues = append(node.PossibleValues, temp)
						} else {
							_, err = outputTable.WriteString(reg.FindStringSubmatch(line)[1] + "\n")
							if err != nil {
								panic(err)
							}
						}
						solFound = true
					}
				}
			}

		} else {
			for {
				line, err = reader.ReadString('\n')
				if err == io.EOF && len(line) == 0 {
					break
				}
				if strings.HasPrefix(line, "v") {
					reg := regexp.MustCompile(".*<values>(.*) </values>.*")
					if SystemSettings.InMemory {
						temp := make([]int, len(node.Variables))
						for i, value := range strings.Split(reg.FindStringSubmatch(line)[1], " ") {
							v, err := strconv.Atoi(value)
							if err != nil {
								panic(err)
							}
							temp[i] = v
						}
						node.PossibleValues = append(node.PossibleValues, temp)
					} else {
						_, err = outputTable.WriteString(reg.FindStringSubmatch(line)[1] + "\n")
						if err != nil {
							panic(err)
						}
					}
					solFound = true
				}
			}
		}
		return solFound
	}

}

/*func doNacreMakeFile(){
	cmd := exec.Command("make")
	cmd.Dir = "/mnt/c/Users/simon/Desktop/Università/Tesi/Programmi/CSP_Project/libs/nacre_master/core"
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}*/

func PrintMemUsage() {
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
}
