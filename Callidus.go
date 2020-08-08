package main

import (
	. "../Callidus/computation"
	. "../Callidus/constraint"
	. "../Callidus/hyperTree"
	. "../Callidus/pre-processing"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	if len(os.Args) == 1 {
		panic("The first parameter must be an xml file or lzma file")
	}
	args := os.Args[1:]
	filePath := args[0]
	if !strings.HasSuffix(filePath, ".xml") && !strings.HasSuffix(filePath, ".lzma") {
		panic("The first parameter must be an xml file or lzma file")
	}

	SystemSettings.InitSettings(args, filePath)

	fmt.Println("Start Callidus")
	start := time.Now()

	fmt.Println("creating hypergraph")
	startTranslation := time.Now()
	HypergraphTranslation(filePath)
	fmt.Println("hypergraph created in ", time.Since(startTranslation))

	hyperTreeRaw := ""
	if SystemSettings.HypertreeFile == "output"+SystemSettings.FolderName+"hypertree" {
		fmt.Println("decomposing hypertree")
		startDecomposition := time.Now()
		hyperTreeRaw = HypertreeDecomposition(filePath, "output"+SystemSettings.FolderName, SystemSettings.InMemory)
		fmt.Println("hypertree decomposed in ", time.Since(startDecomposition))
	}

	var wg sync.WaitGroup
	wg.Add(3)

	fmt.Println("parsing hypertree, domain and constraints")
	startPrep := time.Now()
	var nodes []*Node
	var root *Node
	go func() {
		if SystemSettings.InMemory {
			root, nodes = GetHyperTreeInMemory(&hyperTreeRaw)
		} else {
			root, nodes = GetHyperTree()
		}
		wg.Done()
	}()

	var domains map[string][]int
	go func() {
		domains = GetDomains(filePath)
		wg.Done()
	}()

	var constraints []*Constraint
	go func() {
		constraints = GetConstraints(filePath, "output"+SystemSettings.FolderName)
		wg.Done()
	}()

	wg.Wait()
	fmt.Println("hypertree, domain and constraints parsed in ", time.Since(startPrep))
	if !SystemSettings.Debug {
		err := os.RemoveAll("output" + SystemSettings.FolderName)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("starting sub csp computation")
	/*go func() {
		for {
			PrintMemUsage()
			time.Sleep(5 * time.Second)
		}
	}()*/
	startSubComputation := time.Now()
	satisfiable := SubCSP_Computation(domains, constraints, nodes)
	fmt.Println("sub csp computed in ", time.Since(startSubComputation).Minutes())
	if !satisfiable {
		fmt.Println("NO SOLUTIONS")
		return
	}
	if !SystemSettings.Debug {

		err := os.RemoveAll("subCSP-" + SystemSettings.FolderName)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("starting yannakaki")
	startYannakaki := time.Now()
	Yannakaki(root)
	fmt.Println("yannakaki finished in ", time.Since(startYannakaki))
	fmt.Println("ended in ", time.Since(start))

	if SystemSettings.PrintSol {
		finalResult := make([]map[string]int, 0)
		searchResults(root, &finalResult)
		printSolution(&finalResult)
	}

	if !SystemSettings.Debug {
		err := os.RemoveAll("tables-" + SystemSettings.FolderName)
		if err != nil {
			panic(err)
		}
	}

}

func printSolution(result *[]map[string]int) {
	if len(*result) > 0 {
		if SystemSettings.Output == "" {
			for indexResult, res := range *result {
				fmt.Print("Sol " + strconv.Itoa(indexResult+1) + "\n")
				for key, value := range res {
					fmt.Print(key + " -> " + strconv.Itoa(value) + "\n")
				}
			}
			fmt.Print("Solutions found: " + strconv.Itoa(len(*result)) + "\n")
		} else {
			err := os.RemoveAll(SystemSettings.Output)
			if err != nil {
				panic(err)
			}
			file, err := os.OpenFile(SystemSettings.Output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
			if err != nil {
				panic(err)
			}
			for indexResult, res := range *result {
				_, err = file.WriteString("Sol " + strconv.Itoa(indexResult+1) + "\n")
				if err != nil {
					panic(err)
				}
				for key, value := range res {
					_, err = file.WriteString(key + " -> " + strconv.Itoa(value) + "\n")
					if err != nil {
						panic(err)
					}
				}
			}
			_, err = file.WriteString("Solutions found: " + strconv.Itoa(len(*result)))
			if err != nil {
				panic(err)
			}
		}
	} else {
		fmt.Println("NO SOLUTIONS")
	}
}

func searchResults(actual *Node, finalResults *[]map[string]int) {
	joinVariables := make(map[string]int, 0)
	if actual.Father != nil {
		for _, varFather := range actual.Father.Variables {
			for index, varActual := range actual.Variables {
				if varActual == varFather {
					joinVariables[varActual] = index
					break
				}
			}
		}
	}

	addNewResults := false
	newResults := make([]map[string]int, 0)

	if SystemSettings.InMemory {
		addNewResults, newResults = searchNewResultsInMemory(actual, &joinVariables, finalResults)
	} else {
		addNewResults, newResults = searchNewResultsOnFile(actual, &joinVariables, finalResults)
	}

	if addNewResults {
		for _, singleNewResult := range newResults {
			*finalResults = append(*finalResults, singleNewResult)
		}
	}

	for _, son := range actual.Sons {
		searchResults(son, finalResults)
	}

}

func searchNewResultsInMemory(actual *Node, joinVariables *map[string]int, finalResults *[]map[string]int) (bool, []map[string]int) {
	joinDoneCount := make(map[string]int)
	newResults := make([]map[string]int, 0)
	addNewResults := false

	for _, singleNodeSolution := range actual.PossibleValues {
		computationNewResults(actual, singleNodeSolution, joinVariables, &joinDoneCount, finalResults, &addNewResults, &newResults)
	}

	return addNewResults, newResults
}

func searchNewResultsOnFile(actual *Node, joinVariables *map[string]int, finalResults *[]map[string]int) (bool, []map[string]int) {
	joinDoneCount := make(map[string]int)
	newResults := make([]map[string]int, 0)
	addNewResults := false

	fileActual, rActual := OpenNodeFile(actual.Id)
	for rActual.Scan() {
		singleNodeSolution := GetValues(rActual.Text(), len(actual.Variables))
		computationNewResults(actual, singleNodeSolution, joinVariables, &joinDoneCount, finalResults, &addNewResults, &newResults)
	}

	fileActual.Close()

	return addNewResults, newResults
}

func computationNewResults(actual *Node, singleNodeSolution []int, joinVariables *map[string]int, joinDoneCount *map[string]int, finalResults *[]map[string]int, addNewResults *bool, newResults *[]map[string]int) {
	if singleNodeSolution == nil {
		return
	}

	keyJoin := ""
	for index, value := range singleNodeSolution {
		_, isVariableJoin := (*joinVariables)[actual.Variables[index]]
		if isVariableJoin {
			keyJoin += strconv.Itoa(value)
		}
	}
	_, alreadyInMap := (*joinDoneCount)[keyJoin]
	if alreadyInMap {
		(*joinDoneCount)[keyJoin]++
	} else {
		(*joinDoneCount)[keyJoin] = 1
	}

	if len(*joinVariables) >= 1 {
		for _, singleFinalResult := range *finalResults {
			joinOk := true
			for joinKey, joinIndex := range *joinVariables {
				if singleFinalResult[joinKey] != singleNodeSolution[joinIndex] {
					joinOk = false
					break
				}
			}
			if joinOk {

				if (*joinDoneCount)[keyJoin] == 1 {
					for index, value := range singleNodeSolution {
						singleFinalResult[actual.Variables[index]] = value
					}
				} else {
					*addNewResults = true
					copyRes := make(map[string]int, 0)
					for key, val := range singleFinalResult {
						copyRes[key] = val
					}

					for index, value := range singleNodeSolution {
						copyRes[actual.Variables[index]] = value
					}
					*newResults = append(*newResults, copyRes)
				}

			}
		}
	} else {
		resTemp := make(map[string]int)
		for index, value := range singleNodeSolution {
			resTemp[actual.Variables[index]] = value
		}
		*finalResults = append(*finalResults, resTemp)
	}
}
