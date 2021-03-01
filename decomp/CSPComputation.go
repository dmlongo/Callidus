package decomp

import (
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

func SubCSPComputation(domains map[string][]int, constraints []*Constraint, nodes []*Node, folder string, subDebug bool, inMemory bool, parallel bool) bool {
	subCspFolder := "subCSP-" + folder
	tablesFolder := "tables-" + folder
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
	if parallel {
		subCspSemaphore := &sync.WaitGroup{}
		subCspSemaphore.Add(len(nodes))
		checkSatisfiableSemaphore := &sync.WaitGroup{}
		checkSatisfiableSemaphore.Add(1)
		satisfiableChan := make(chan bool, len(nodes))
		for _, node := range nodes {
			go CreateAndSolveSubCSP(subCspFolder, tablesFolder, node, domains, constraints, subCspSemaphore, satisfiableChan, subDebug, inMemory, parallel)
			// CreateAndSolveSubCSP(subCspFolder, tablesFolder, node, domains, constraints, subCspSemaphore, satisfiableChan, subDebug, inMemory, parallel)
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
			satisfiable = CreateAndSolveSubCSP(subCspFolder, tablesFolder, node, domains, constraints, nil, nil, subDebug, inMemory, parallel)
			if !satisfiable {
				break
			}
		}
	}
	return satisfiable
}

func getPossibleValues(constraint *Constraint) string {
	possibleValues := "<supports> "
	if !constraint.CType {
		possibleValues = "<conflicts> "
	}
	for _, tup := range constraint.Relation {
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

func solve(xmlFile string, tableFile string, satisfiableChan chan bool, node *Node, subDebug bool, inMemory bool, parallel bool) bool {
	defer func(debugOption bool) {
		if !debugOption {
			err := os.Remove(xmlFile)
			if err != nil {
				panic(err)
			}
		}
	}(subDebug)
	cmd := exec.Command("./libs/nacre", xmlFile, "-complete", "-sols", "-verb=3")
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
	if inMemory {
		node.PossibleValues = make([][]int, 0)
	} else {
		outputTable, err = os.OpenFile(tableFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	}
	if err != nil {
		panic(err)
	}
	if parallel {
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
					if inMemory {
						temp := make([]int, len(node.Bag))
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
				if inMemory {
					temp := make([]int, len(node.Bag))
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
