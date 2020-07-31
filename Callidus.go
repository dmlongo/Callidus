package main

import (
	. "../Callidus/computation"
	. "../Callidus/constraint"
	. "../Callidus/hyperTree"
	. "../Callidus/pre-processing"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type valueSol struct {
	value    int
	whoAdded int
}

func main() {
	if len(os.Args) == 1 {
		panic("The first parameter must be an xml file or lzma file")
	}
	args := os.Args[1:]
	filePath := args[0]
	if !strings.HasSuffix(filePath, ".xml") && !strings.HasSuffix(filePath, ".lzma") {
		panic("The first parameter must be an xml file or lzma file")
	}
	yannakakiVersion := selectYannakakiVersion(args) //true if parallel, false if sequential
	inMemory := selectComputation(args)              // -i for computation in memory, inMemory = true or inMemory = false if not
	debugOption := selectDebugOption(args)
	parallelSubComputation := selectSubComputationExec(args)
	balancedAlgorithm := selectBalancedAlgorithm(args)
	printSol := selectPrintSol(args)
	computeWidth := selectComputeWidth(args)
	outputFile := writeSolution(args)

	fmt.Println("Start Callidus")
	start := time.Now()

	fmt.Println("creating hypergraph")
	startTranslation := time.Now()
	HypergraphTranslation(filePath)
	fmt.Println("hypergraph created in ", time.Since(startTranslation))
	folderName := getFolderName(filePath)

	hypertreeFile := takeHypertreeFile(args, "output"+folderName)
	hyperTreeRaw := ""
	if hypertreeFile == "output"+folderName+"hypertree" {
		fmt.Println("decomposing hypertree")
		startDecomposition := time.Now()
		hyperTreeRaw = HypertreeDecomposition(filePath, "output"+folderName, balancedAlgorithm, inMemory, computeWidth)
		fmt.Println("hypertree decomposed in ", time.Since(startDecomposition))
	}

	var wg sync.WaitGroup
	wg.Add(3)

	fmt.Println("parsing hypertree, domain and constraints")
	startPrep := time.Now()
	var nodes []*Node
	var root *Node
	go func() {
		if inMemory {
			root, nodes = GetHyperTreeInMemory(&hyperTreeRaw)
		} else {
			root, nodes = GetHyperTree(hypertreeFile)
		}
		wg.Done()
	}()

	var domains map[string][]int
	go func() {
		domains = GetDomains(filePath, "output"+folderName)
		wg.Done()
	}()

	var constraints []*Constraint
	go func() {
		constraints = GetConstraints(filePath, "output"+folderName)
		wg.Done()
	}()

	wg.Wait()
	fmt.Println("hypertree, domain and constraints parsed in ", time.Since(startPrep))
	if !debugOption {
		err := os.RemoveAll("output" + folderName)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("starting sub csp computation")
	startSubComputation := time.Now()
	satisfiable := SubCSP_Computation("subCSP-"+folderName, domains, constraints, nodes, parallelSubComputation, debugOption)
	fmt.Println("sub csp computed in ", time.Since(startSubComputation))
	if !satisfiable {
		fmt.Println("NO SOLUTIONS")
		return
	}
	if !debugOption {
		err := os.RemoveAll("subCSP-" + folderName)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("starting yannakaki")
	startYannakaki := time.Now()
	Yannakaki(root, yannakakiVersion)
	fmt.Println("yannakaki finished in ", time.Since(startYannakaki))
	fmt.Println("ended in ", time.Since(start))

	if printSol {
		//printSolution(root, outputFile, len(domains))
		//Simone's version

		result := make([]map[string]int, 0)
		result = *(searchResults(root, &result))
		fmt.Println(len(result))
		printSolution2(&result, outputFile)
	}
}

func printSolution2(result *[]map[string]int, outputFile string) {
	if len(*result) > 0 {
		if outputFile == "" {
			//fmt.Println("SOLUTIONS FOUND: " + strconv.Itoa(len(result)) + "\n")
			//TODO: fare il prodotto cartesiano
		} else {
			err := os.RemoveAll(outputFile)
			if err != nil {
				panic(err)
			}
			file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
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

func searchResults(actual *Node, result *[]map[string]int) *[]map[string]int {
	joins := make(map[string]int, 0)
	if actual.Father != nil {
		for _, varFather := range actual.Father.Variables {
			for index, varActual := range actual.Variables {
				if varActual == varFather {
					joins[varActual] = index
					break
				}
			}
		}
	}
	joinDone := make(map[string]int)
	newResults := make([]map[string]int, 0)
	addNewResults := false

	for _, sol := range actual.PossibleValues {

		keyJoin := ""
		for index, value := range sol {
			_, isVariableJoin := joins[actual.Variables[index]]
			if isVariableJoin {
				keyJoin += strconv.Itoa(value)
			}
		}
		_, alreadyInMap := joinDone[keyJoin]
		if alreadyInMap {
			joinDone[keyJoin]++
		} else {
			joinDone[keyJoin] = 1
		}

		if len(joins) >= 1 {
			for _, res := range *result {
				joinOk := true
				for joinKey, joinIndex := range joins {
					if res[joinKey] != sol[joinIndex] {
						joinOk = false
						break
					}
				}
				if joinOk {

					if joinDone[keyJoin] == 1 {
						for index, value := range sol {
							res[actual.Variables[index]] = value
						}
					} else {
						addNewResults = true
						copyRes := make(map[string]int, 0)
						for key, val := range res {
							copyRes[key] = val
						}

						for index, value := range sol {
							copyRes[actual.Variables[index]] = value
						}
						newResults = append(newResults, copyRes)
					}

				}
			}
		} else {
			resTemp := make(map[string]int)
			for index, value := range sol {
				resTemp[actual.Variables[index]] = value
			}
			*result = append(*result, resTemp)
		}

	}

	if addNewResults {
		for _, singleNewResult := range newResults {
			*result = append(*result, singleNewResult)
		}
	}

	for _, son := range actual.Sons {
		result = searchResults(son, result)
	}
	return result
}

func printSolution(root *Node, outputFile string, numVariables int) {
	if outputFile == "" {
		contSol := 0
		sol := make(map[string]valueSol)
		printAllSolutions(root, sol, &contSol, numVariables, nil)
		fmt.Println("Solutions found: ", contSol)
	} else {
		file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
		if err != nil {
			panic(err)
		}
		contSol := 0
		sol := make(map[string]valueSol)
		printAllSolutions(root, sol, &contSol, numVariables, file)
		_, err = file.WriteString("Solutions found: " + strconv.Itoa(contSol))
		if err != nil {
			panic(err)
		}

	}
}

func printAllSolutions(node *Node, sol map[string]valueSol, contSol *int, numVariables int, outputFile *os.File) {
	x := 0
	for x < len(node.PossibleValues) {
		if canAddToSol(node, sol, x) {
			if len(sol) == numVariables {
				*contSol++
				printSol(sol, outputFile, contSol)
			}
			for _, son := range node.Sons {
				printAllSolutions(son, sol, contSol, numVariables, outputFile)
			}
			removeNodeFromSol(node, sol)
			x++
		} else {
			x++
		}
	}
}

func canAddToSol(node *Node, sol map[string]valueSol, row int) bool {
	for index, variable := range node.Variables {
		if _, exist := sol[variable]; exist {
			if node.PossibleValues[row][index] != sol[variable].value {
				return false
			}
		} else {
			sol[variable] = valueSol{value: node.PossibleValues[row][index], whoAdded: node.Id}
		}
	}
	return true
}

func removeNodeFromSol(node *Node, sol map[string]valueSol) {
	for _, variable := range node.Variables {
		if sol[variable].whoAdded == node.Id {
			delete(sol, variable)
		}
	}
}

func printSol(sol map[string]valueSol, outputFile *os.File, numSol *int) {
	solString := "Sol " + strconv.Itoa(*numSol) + "\n"
	for variable, value := range sol {
		solString += variable + " -> " + strconv.Itoa(value.value) + "\n"
	}
	if outputFile != nil {
		_, err := outputFile.WriteString(solString)
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Print(solString)
	}
}

func contains(args []string, param string) int {
	for i := 0; i < len(args); i++ {
		if args[i] == param {
			return i
		}
	}
	return -1
}

func takeHypertreeFile(args []string, folderName string) string {
	i := contains(args, "-h")
	if i == -1 {
		i = contains(args, "--hypertree")
	}
	if i != -1 {
		return args[i+1]
	}
	return folderName + "hypertree"
}

func selectYannakakiVersion(args []string) bool {
	i := contains(args, "-y")
	if i == -1 {
		i = contains(args, "--yannakaki")
	}
	if i != -1 {
		version := args[i+1]
		if version != "s" && version != "p" {
			panic(args[i] + " must be followed by 's' or 'p'")
		}
		if version == "p" {
			return true
		}
	}
	return false
}

func selectComputation(args []string) bool {
	i := contains(args, "-i")
	if i != -1 {
		return true
	}
	return false
}

func selectDebugOption(args []string) bool {
	if contains(args, "-d") != -1 || contains(args, "--debug") != -1 {
		return true
	}
	return false
}

func selectSubComputationExec(args []string) bool {
	if i := contains(args, "-sc"); i != -1 {
		if args[i+1] != "p" && args[i+1] != "s" {
			panic(args[i] + " must be followed by 's' or 'p'")
		}
		if args[i+1] == "s" {
			return false
		}
	}
	return true
}

func selectBalancedAlgorithm(args []string) string {
	i := contains(args, "-b")
	if i == -1 {
		i = contains(args, "--balanced")
	}
	if i != -1 {
		if args[i+1] != "det" && args[i+1] != "balDet" {
			panic(args[i] + " must be followed by 'det' or 'balDet'")
		}
		return args[i+1]
	}
	return "det"
}

func selectPrintSol(args []string) bool {
	if i := contains(args, "-printSol"); i != -1 {
		if args[i+1] != "yes" && args[i+1] != "no" {
			panic(args[i] + " must be followed by 'yes' or 'no'")
		}
		if args[i+1] == "no" {
			return false
		}
	}
	return true
}

func getFolderName(filePath string) string {
	re := regexp.MustCompile(".*/")
	folderName := re.ReplaceAllString(filePath, "")
	re = regexp.MustCompile("\\..*")
	folderName = re.ReplaceAllString(folderName, "")
	folderName = folderName + "/"
	return folderName
}

func selectComputeWidth(args []string) bool {
	if i := contains(args, "-computeWidth"); i != -1 {
		if args[i+1] != "yes" && args[i+1] != "no" {
			panic(args[i] + " must be followed by 'yes' or 'no'")
		}
		if args[i+1] == "no" {
			return false
		}
	}
	return true
}

func writeSolution(args []string) string {
	if i := contains(args, "-output"); i != -1 {
		return args[i+1]
	}
	return ""
}
