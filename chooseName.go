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

func main() {

	args := os.Args[1:]
	fmt.Println(args)
	filePath := args[0]
	if !strings.HasSuffix(filePath, ".xml") {
		panic("The first parameter must be an xml file")
	}
	hypertreeFile := takeHypertreeFile(args)

	fmt.Println("Start")

	fmt.Println("creating hypergraph")
	HypergraphTranslation(filePath)
	fmt.Println("hypergraph created")

	if hypertreeFile == "output/hypertree" {
		fmt.Println("hypertree decomposition")
		HypertreeDecomposition(filePath)
		fmt.Println("hypertree ready")
	}

	var wg sync.WaitGroup
	wg.Add(3)

	var nodes []*Node
	var root *Node
	go func() {
		fmt.Println("creating hypertree")
		root, nodes = GetHyperTree(hypertreeFile)
		fmt.Println("hypertree created")
		wg.Done()
	}()

	var domains map[string][]int
	go func() {
		fmt.Println("creating domain file")
		domains = GetDomains(filePath)
		fmt.Println("domain file created")
		wg.Done()
	}()

	var constraints []*Constraint
	go func() {
		fmt.Println("reading constraints")
		constraints = GetConstraints(filePath)
		fmt.Println("constraints taken")
		wg.Done()
	}()

	wg.Wait()
	err := os.RemoveAll("output")
	if err != nil {
		panic(err)
	}

	fmt.Println("starting sub csp computation")
	SubCSP_Computation(domains, constraints, nodes)
	fmt.Println("sub csp computed")
	//start := time.Now()
	fmt.Println("adding tables to nodes")
	AttachPossibleSolutions(nodes)
	fmt.Println("tables added")
	err = os.RemoveAll("subCSP")
	if err != nil {
		panic(err)
	}

	start := time.Now()
	fmt.Println("starting yannakaki")
	ParallelYannakaki(root)
	fmt.Println("yannakaki finished")
	fmt.Println(time.Since(start).Milliseconds())
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
	if i := contains(args, "-h"); i != -1 {
		return args[i+1]
	}
	if i := contains(args, "--hypertree"); i != -1 {
		return args[i+1]
	}
	return "output/hypertree"
}
