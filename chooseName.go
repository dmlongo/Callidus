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
	filePath := args[0]
	if !strings.HasSuffix(filePath, ".xml") {
		panic("The first parameter must be an xml file")
	}
	hypertreeFile := takeHypertreeFile(args)
	yannakakiVersion := selectYannakakiVersion(args) //true if parallale, false if sequential
	fmt.Println(yannakakiVersion)
	fmt.Println("Start")
	start := time.Now()

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
	/*err := os.RemoveAll("output")
	if err != nil {
		panic(err)
	}*/

	fmt.Println("starting sub csp computation")
	SubCSP_Computation(domains, constraints, nodes)
	fmt.Println("sub csp computed")
	fmt.Println("adding tables to nodes")
	satisfiable := AttachPossibleSolutions(nodes)
	if !satisfiable {
		fmt.Println("NO SOLUTIONS")
		return
	}
	fmt.Println("tables added")
	/*err = os.RemoveAll("subCSP")
	if err != nil {
		panic(err)
	}*/

	fmt.Println("starting yannakaki")
	Yannakaki(root, yannakakiVersion)
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
