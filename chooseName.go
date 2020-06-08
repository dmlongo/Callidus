package main

import (
	. "../CSP_Project/computation"
	. "../CSP_Project/constraint"
	. "../CSP_Project/hyperTree"
	. "../CSP_Project/pre-processing"
	"os"
	"strings"
	"sync"
)

func main() {

	filePath := os.Args[1]
	if !strings.HasSuffix(filePath, ".xml") {
		panic("File must be an xml")
	}

	HypergraphTranslation(filePath)
	HypertreeDecomposition(filePath)

	var wg sync.WaitGroup
	wg.Add(3)

	var nodes []*Node
	go func() {
		_, nodes = GetHyperTree() //TODO: NON FATE GLI STRONZI
		wg.Done()
	}()

	var domains map[string][]int
	go func() {
		domains = GetDomains(filePath)
		wg.Done()
	}()

	var constraints []*Constraint
	go func() {
		constraints = GetConstraints(filePath)
		wg.Done()
	}()

	wg.Wait()

	SubCSP_Computation(domains, constraints, nodes)
	//start := time.Now()
	AttachPossibleSolutions(nodes)
	//fmt.Println(time.Since(start))

}
