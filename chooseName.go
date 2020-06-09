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
	var root *Node
	go func() {
		root, nodes = GetHyperTree()
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
	/*d := &Node{Id: 1, Variables: []string{"Y", "P"}, PossibleValues: [][]int{{3,8},{3,7},{5,7},{6,7}}}
	r := &Node{Id: 2, Variables: []string{"Y", "Z", "U"}, PossibleValues: [][]int{{3,8,9},{9,3,8},{8,3,8},{3,8,4},{3,8,3},{8,9,4},{9,4,7}}}
	s := &Node{Id: 3, Variables: []string{"Z", "U", "W"}, PossibleValues: [][]int{{3,8,9},{9,3,8},{8,3,8},{3,8,4},{3,8,3},{8,9,4},{9,4,7}}}
	t := &Node{Id: 4, Variables: []string{"V", "Z"}, PossibleValues: [][]int{{9,8},{9,3},{9,5}}}
	d.AddSon(r)
	r.AddFather(d)
	r.AddSon(s)
	s.AddFather(r)
	r.AddSon(t)
	t.AddFather(r)*/
	SequentialYannakaki(root)
}
