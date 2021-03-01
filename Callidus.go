package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/dmlongo/callidus/decomp"
	"github.com/dmlongo/callidus/ext"
)

var csp, ht, out string
var subInMem, subSeq, subDebug, ySeq, printSol bool
var folder string
var contSols int

func main() {
	setFlags()

	fmt.Println("Callidus starts!")
	start := time.Now()

	fmt.Print("Creating hypergraph... ")
	startConversion := time.Now()
	ext.Convert(csp)
	fmt.Println("done in", time.Since(startConversion))

	var hyperTreeRaw string
	if ht == "" {
		fmt.Print("Decomposing hypergraph... ")
		startDecomposition := time.Now()
		hyperTreeRaw = ext.Decompose(csp, "output"+folder, subInMem)
		fmt.Println("done in", time.Since(startDecomposition))
		ht = "output" + folder + "hypertree"
	}

	var wg sync.WaitGroup
	wg.Add(3)

	fmt.Print("Parsing hypertree, domains and constraints... ")
	startPrep := time.Now()
	var nodes []*decomp.Node
	var root *decomp.Node
	go func() {
		if subInMem {
			root, nodes = ext.ParseDecompInMemory(&hyperTreeRaw)
		} else {
			root, nodes = ext.ParseDecomp(ht)
		}
		wg.Done()
	}()

	var domains map[string][]int
	go func() {
		domains = ext.ParseDomains(csp, folder)
		wg.Done()
	}()

	var constraints []*decomp.Constraint
	go func() {
		constraints = ext.ParseConstraints(csp, "output"+folder)
		wg.Done()
	}()

	wg.Wait()
	fmt.Println("done in", time.Since(startPrep))
	if !subDebug {
		err := os.RemoveAll("output" + folder)
		if err != nil {
			panic(err)
		}
	}

	fmt.Print("Solving sub-CSPs... ")
	//go func() {
	//	for {
	//		PrintMemUsage()
	//		time.Sleep(5 * time.Second)
	//	}
	//}()
	startSubComputation := time.Now()
	satisfiable := decomp.SubCSPComputation(domains, constraints, nodes, folder, subDebug, subInMem, !subSeq)
	fmt.Println("done in", time.Since(startSubComputation))
	if !satisfiable {
		fmt.Println("NO SOLUTIONS")
		return
	}
	if !subDebug {
		err := os.RemoveAll("subCSP-" + folder)
		if err != nil {
			panic(err)
		}
	}
	fmt.Print("Running Yannakakis... ")
	startYannakakis := time.Now()
	decomp.Yannakakis(root, !subSeq, subInMem, folder)
	fmt.Println("done in", time.Since(startYannakakis))
	fmt.Println("Callidus solved", csp, "in", time.Since(start))

	finalResult := make([]map[string]int, 0)

	startSearchResult := time.Now()
	searchResults(root, &finalResult)
	fmt.Println("Search results ended in", time.Since(startSearchResult))

	if printSol {
		printSolution(&finalResult)
	}

	if !subDebug {
		err := os.RemoveAll("tables-" + folder)
		if err != nil {
			panic(err)
		}
	}
}

func setFlags() {
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	flagSet.SetOutput(ioutil.Discard) //todo: see what happens without this line

	flagSet.StringVar(&csp, "csp", "", "Path to the CSP to solve (XCSP3 format)")
	flagSet.StringVar(&ht, "ht", "", "Path to a decomposition of the CSP to solve (GML format)")
	flagSet.StringVar(&out, "out", "", "Save the solutions of the CSP into the specified file")
	flagSet.BoolVar(&subInMem, "subInMem", false, "Activate in-memory computation of sub-CSPs")
	flagSet.BoolVar(&subSeq, "subSeq", false, "Activate sequential computation of sub-CSPs")
	flagSet.BoolVar(&subDebug, "subDebug", false, "Activate debug for sub-CSPs")
	flagSet.BoolVar(&ySeq, "ySeq", false, "Use sequential Yannakakis' algorithm")
	flagSet.BoolVar(&printSol, "printSol", true, "Print solutions of the CSP")

	parseError := flagSet.Parse(os.Args[1:])
	if parseError != nil {
		fmt.Print("Parse Error:\n", parseError.Error(), "\n\n")
	}

	if parseError != nil || csp == "" {
		out := fmt.Sprint("Usage of Callidus (https://github.com/dmlongo/Callidus)\n")
		flagSet.VisitAll(func(f *flag.Flag) {
			if f.Name != "csp" {
				return
			}
			s := fmt.Sprintf("%T", f.Value) // used to get type of flag
			if s[6:len(s)-5] != "bool" {
				out += fmt.Sprintf("  -%-10s \t<%s>\n", f.Name, s[6:len(s)-5])
			} else {
				out += fmt.Sprintf("  -%-10s \n", f.Name)
			}
			out += fmt.Sprintln("\t" + f.Usage)
		})
		out += fmt.Sprintln("\nOptional Arguments: ")
		flagSet.VisitAll(func(f *flag.Flag) {
			if f.Name == "csp" {
				return
			}
			s := fmt.Sprintf("%T", f.Value) // used to get type of flag
			if s[6:len(s)-5] != "bool" {
				out += fmt.Sprintf("  -%-10s \t<%s>\n", f.Name, s[6:len(s)-5])
			} else {
				out += fmt.Sprintf("  -%-10s \n", f.Name)
			}
			out += fmt.Sprintln("\t" + f.Usage)
		})
		fmt.Fprintln(os.Stderr, out)

		os.Exit(1)
	}

	re := regexp.MustCompile(".*/")
	folder = re.ReplaceAllString(csp, "")
	re = regexp.MustCompile("\\..*")
	folder = re.ReplaceAllString(folder, "")
	folder = folder + "/"

	//if ht == "" {ht = "output" + folder + "hypertree"}
	//fmt.Printf("csp=%v\nht=%v\nout=%v\n", csp, ht, out)

	contSols = 0
}

func printSolution(result *[]map[string]int) {
	if len(*result) > 0 {
		if out == "" {
			for indexResult, res := range *result {
				fmt.Print("Sol " + strconv.Itoa(indexResult+1) + "\n")
				for key, value := range res {
					fmt.Print(key + " -> " + strconv.Itoa(value) + "\n")
				}
			}
			fmt.Print("Solutions found: " + strconv.Itoa(len(*result)) + "\n")
		} else {
			err := os.RemoveAll(out)
			if err != nil {
				panic(err)
			}
			file, err := os.OpenFile(out, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
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

func searchResults(actual *decomp.Node, finalResults *[]map[string]int) {
	joinVariables := make(map[string]int, 0)
	if actual.Father != nil {
		for _, varFather := range actual.Father.Bag {
			for index, varActual := range actual.Bag {
				if varActual == varFather {
					joinVariables[varActual] = index
					break
				}
			}
		}
	}

	addNewResults := false
	newResults := make([]map[string]int, 0)

	if subInMem {
		addNewResults, newResults = searchNewResultsInMemory(actual, &joinVariables, finalResults)
	} else {
		addNewResults, newResults = searchNewResultsOnFile(actual, &joinVariables, finalResults)
		if newResults == nil {
			*finalResults = nil
			return
		}
	}

	if addNewResults {
		for _, singleNewResult := range newResults {
			*finalResults = append(*finalResults, singleNewResult)
		}
	}

	fmt.Println(len(*finalResults))

	for _, son := range actual.Children {
		searchResults(son, finalResults)
	}

}

func searchNewResultsInMemory(actual *decomp.Node, joinVariables *map[string]int, finalResults *[]map[string]int) (bool, []map[string]int) {
	joinDoneCount := make(map[string]int)
	newResults := make([]map[string]int, 0)
	addNewResults := false

	for _, singleNodeSolution := range actual.PossibleValues {
		computationNewResults(actual, singleNodeSolution, joinVariables, &joinDoneCount, finalResults, &addNewResults, &newResults)
	}

	return addNewResults, newResults
}

func searchNewResultsOnFile(actual *decomp.Node, joinVariables *map[string]int, finalResults *[]map[string]int) (bool, []map[string]int) {
	joinDoneCount := make(map[string]int)
	newResults := make([]map[string]int, 0)
	addNewResults := false

	fileActual, rActual := decomp.OpenNodeFile(actual.ID, folder)
	for rActual.Scan() {
		singleNodeSolution := decomp.GetValues(rActual.Text(), len(actual.Bag))
		if singleNodeSolution == nil {
			break
		}
		if singleNodeSolution[0] == -1 {
			return false, nil
		}
		computationNewResults(actual, singleNodeSolution, joinVariables, &joinDoneCount, finalResults, &addNewResults, &newResults)
	}

	fileActual.Close()

	return addNewResults, newResults
}

func computationNewResults(actual *decomp.Node, singleNodeSolution []int, joinVariables *map[string]int, joinDoneCount *map[string]int, finalResults *[]map[string]int, addNewResults *bool, newResults *[]map[string]int) {
	if singleNodeSolution == nil {
		return
	}

	keyJoin := ""
	for index, value := range singleNodeSolution {
		_, isVariableJoin := (*joinVariables)[actual.Bag[index]]
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
						singleFinalResult[actual.Bag[index]] = value
					}
				} else {
					*addNewResults = true
					copyRes := make(map[string]int, 0)
					for key, val := range singleFinalResult {
						copyRes[key] = val
					}

					for index, value := range singleNodeSolution {
						copyRes[actual.Bag[index]] = value
					}
					*newResults = append(*newResults, copyRes)
				}

			}
		}
	} else {
		resTemp := make(map[string]int)
		for index, value := range singleNodeSolution {
			resTemp[actual.Bag[index]] = value
		}
		*finalResults = append(*finalResults, resTemp)
	}
}
