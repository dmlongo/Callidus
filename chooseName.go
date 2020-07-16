package main

import (
	. "../CSP_Project/computation"
	. "../CSP_Project/constraint"
	. "../CSP_Project/hyperTree"
	. "../CSP_Project/pre-processing"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type Result map[string]int

func main() {
	args := os.Args[1:]
	filePath := args[0]
	if !strings.HasSuffix(filePath, ".xml") && !strings.HasSuffix(filePath, ".lzma") {
		panic("The first parameter must be an xml file or lzma file")
	}
	hypertreeFile := takeHypertreeFile(args)
	yannakakiVersion := selectYannakakiVersion(args) //true if parallel, false if sequential
	inMemory := selectComputation(args)              // -i for computation in memory, inMemory = true or inMemory = false if not
	solver := selectSolver(args)
	debugOption := selectDebugOption(args)
	parallelSubComputation := selectSubComputationExec(args)
	balancedAlgorithm := selectBalancedAlgorithm(args)

	fmt.Println("Start")
	start := time.Now()

	fmt.Println("creating hypergraph")
	startTranslation := time.Now()
	HypergraphTranslation(filePath)
	fmt.Println("hypergraph in ", time.Since(startTranslation))

	hyperTreeRaw := ""
	if hypertreeFile == "output/hypertree" {
		fmt.Println("decomposing hypertree")
		startDecomposition := time.Now()
		hyperTreeRaw = HypertreeDecomposition(filePath, balancedAlgorithm, inMemory)
		fmt.Println("hypertree decomposed in ", time.Since(startDecomposition))
	}

	var wg sync.WaitGroup
	wg.Add(3)

	var nodes []*Node
	var root *Node
	go func() {
		fmt.Println("creating hypertree")
		startHypertreeCreation := time.Now()
		if inMemory {
			root, nodes = GetHyperTreeInMemory(hypertreeFile, &hyperTreeRaw)
		} else {
			root, nodes = GetHyperTree(hypertreeFile)
		}
		fmt.Println("hypertree created in ", time.Since(startHypertreeCreation))
		wg.Done()
	}()

	var domains map[string][]int
	var variables []string
	go func() {
		fmt.Println("parsing domain")
		startDomainParsing := time.Now()
		domains, variables = GetDomains(filePath)
		fmt.Println("domain parsed in ", time.Since(startDomainParsing))
		wg.Done()
	}()

	var constraints []*Constraint
	go func() {
		fmt.Println("reading constraints")
		startConstraintsParsing := time.Now()
		constraints = GetConstraints(filePath)
		fmt.Println("constraints taken in ", time.Since(startConstraintsParsing))
		wg.Done()
	}()

	wg.Wait()
	if !debugOption {
		err := os.RemoveAll("output")
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("starting sub csp computation")
	startSubComputation := time.Now()
	solutions := SubCSP_Computation(domains, constraints, nodes, inMemory, solver, parallelSubComputation)
	fmt.Println("sub csp computed in ", time.Since(startSubComputation))

	fmt.Println("adding tables to nodes")
	startAddingTables := time.Now()
	satisfiable := AttachPossibleSolutions(nodes, &solutions, inMemory, solver)
	if !satisfiable {
		fmt.Println("NO SOLUTIONS")
		return
	}
	fmt.Println("tables added in ", time.Since(startAddingTables))

	if !debugOption {
		err := os.RemoveAll("subCSP")
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("starting yannakaki")
	startYannakaki := time.Now()
	Yannakaki(root, yannakakiVersion)
	fmt.Println("yannakaki finished in ", time.Since(startYannakaki))
	fmt.Println("ended in ", time.Since(start))

	var results []*Result
	for i, node := range nodes {
		if i == 0 {
			for _, arrayNodeSingleResult := range node.PossibleValues {
				res := make(Result)
				for indexVariable, singleValue := range arrayNodeSingleResult {
					res[node.Variables[indexVariable]] = singleValue
				}
				results = append(results, &res)
			}
		} else {
			for indexResult, arrayNodeSingleResult := range node.PossibleValues {
				res := results[indexResult]
				for indexVariable, singleValue := range arrayNodeSingleResult {
					(*res)[node.Variables[indexVariable]] = singleValue
				}
			}
		}
	}
	if len(results) > 0 {
		fmt.Println("SOLUTIONS:")
		for _, result := range results {
			fmt.Println(result)
		}
	} else {
		fmt.Println("NO SOLUTIONS")
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

func takeHypertreeFile(args []string) string {
	i := contains(args, "-h")
	if i == -1 {
		i = contains(args, "--hypertree")
	}
	if i != -1 {
		return args[i+1]
	}
	return "output/hypertree"
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
		if version == "s" {
			return false
		}
	}
	return true
}

func selectComputation(args []string) bool {
	i := contains(args, "-i")
	if i != -1 {
		return true
	}
	return false
}

func selectSolver(args []string) string {
	i := contains(args, "-s")
	if i == -1 {
		i = contains(args, "--solver")
	}
	if i != -1 {
		return args[i+1]
	}
	return "Nacre"
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
	return "balDet"
}
