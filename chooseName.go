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

	filePath := os.Args[1]
	if !strings.HasSuffix(filePath, ".xml") {
		panic("File must be an xml")
	}

	fmt.Println("Start")

	fmt.Println("creating hypergraph")
	HypergraphTranslation(filePath)
	fmt.Println("hypergraph created")

	/*fmt.Println("hypertree decomposition")
	HypertreeDecomposition(filePath)
	fmt.Println("hypertree ready")*/

	var wg sync.WaitGroup
	wg.Add(3)

	var nodes []*Node
	var root *Node
	go func() {
		fmt.Println("creating hypertree")
		root, nodes = GetHyperTree()
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

	fmt.Println("starting sub csp computation")
	SubCSP_Computation(domains, constraints, nodes)
	fmt.Println("sub csp computed")
	//start := time.Now()
	fmt.Println("adding tables to nodes")
	AttachPossibleSolutions(nodes)
	fmt.Println("tables added")

	start := time.Now()
	fmt.Println("starting yannakaki")
	ParallelYannakaki(root)
	fmt.Println("yannakaki finished")
	fmt.Println(time.Since(start).Milliseconds())
	/*for _, node := range nodes {
		fmt.Println(node)
	}*/
}
