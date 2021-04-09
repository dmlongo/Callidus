package decomp

import (
	"sync"
	"testing"

	"github.com/dmlongo/callidus/ctr"
)

func TestSeq1(t *testing.T) {
	input, partial, output, sols := test1Data()
	if !equals(YannakakisSeq(input), partial) {
		t.Error("y(input) != partial")
	}
	if !equals(FullyReduceRelationsSeq(partial), output) {
		t.Error("y(partial) != output")
	}
	if !solEquals(ComputeAllSolutions(output), sols) {
		t.Error("y(output) != solutions")
	}
}

func TestSeq2(t *testing.T) {
	input, partial, output, sols := test2Data()
	if !equals(YannakakisSeq(input), partial) {
		t.Error("y(input) != partial")
	}
	if !equals(FullyReduceRelationsSeq(partial), output) {
		t.Error("y(partial) != output")
	}
	if !solEquals(ComputeAllSolutions(output), sols) {
		t.Error("y(output) != solutions")
	}
}

func TestPar1(t *testing.T) {
	input, partial, output, sols := test1Data()
	if !equals(YannakakisPar(input), partial) {
		t.Error("y(input) != partial")
	}
	if !equals(FullyReduceRelationsPar(partial), output) {
		t.Error("y(partial) != output")
	}
	if !solEquals(ComputeAllSolutions(output), sols) {
		t.Error("y(output) != solutions")
	}
}

func TestPar2(t *testing.T) {
	input, partial, output, sols := test2Data()
	if !equals(YannakakisPar(input), partial) {
		t.Error("y(input) != partial")
	}
	if !equals(FullyReduceRelationsPar(partial), output) {
		t.Error("y(partial) != output")
	}
	if !solEquals(ComputeAllSolutions(output), sols) {
		t.Error("y(output) != solutions")
	}
}

func equals(node1 *Node, node2 *Node) bool {
	if !samePossibleValues(node1, node2) {
		return false
	}
	for i := range node1.Children {
		if !equals(node1.Children[i], node2.Children[i]) {
			return false
		}
	}
	return true
}

func solEquals(sols1 []ctr.Solution, sols2 []ctr.Solution) bool {
	return subsetOf(sols1, sols2) && subsetOf(sols2, sols1)
}

func subsetOf(sols1 []ctr.Solution, sols2 []ctr.Solution) bool {
	for _, s1 := range sols1 {
		found := false
		for _, s2 := range sols2 {
			if s1.Equals(s2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func samePossibleValues(node1 *Node, node2 *Node) bool {
	if node1.ID != node2.ID {
		return false
	}
	if len(node1.Tuples) != len(node2.Tuples) {
		return false
	}
	for i := range node1.Tuples {
		if len(node1.Tuples[i]) != len(node2.Tuples[i]) {
			return false
		}
		for j := range node1.Tuples[i] {
			if node1.Tuples[i][j] != node2.Tuples[i][j] {
				return false
			}
		}
	}
	return true
}

func test1Data() (*Node, *Node, *Node, []ctr.Solution) {
	//creating input
	dInput := &Node{ID: 1, Tuples: Relation{{3, 8}, {3, 7}, {5, 7}, {6, 7}}, Lock: &sync.Mutex{}}
	dInput.SetBag([]string{"Y", "P"})
	rInput := &Node{ID: 2, Tuples: Relation{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 4}, {3, 8, 3}, {8, 9, 4}, {9, 4, 7}}, Lock: &sync.Mutex{}}
	rInput.SetBag([]string{"Y", "Z", "U"})
	sInput := &Node{ID: 3, Tuples: Relation{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 4}, {3, 8, 3}, {8, 9, 4}, {9, 4, 7}}, Lock: &sync.Mutex{}}
	sInput.SetBag([]string{"Z", "U", "W"})
	tInput := &Node{ID: 4, Tuples: Relation{{9, 8}, {9, 3}, {9, 5}}, Lock: &sync.Mutex{}}
	tInput.SetBag([]string{"V", "Z"})
	dInput.AddChild(rInput)
	rInput.AddChild(sInput)
	rInput.AddChild(tInput)

	// creating partially reduced
	dPartial := &Node{ID: 1, Tuples: Relation{{3, 8}, {3, 7}}, Lock: &sync.Mutex{}}
	dPartial.SetBag([]string{"Y", "P"})
	rPartial := &Node{ID: 2, Tuples: Relation{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 3}}, Lock: &sync.Mutex{}}
	rPartial.SetBag([]string{"Y", "Z", "U"})
	sPartial := &Node{ID: 3, Tuples: Relation{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 4}, {3, 8, 3}, {8, 9, 4}, {9, 4, 7}}, Lock: &sync.Mutex{}}
	sPartial.SetBag([]string{"Z", "U", "W"})
	tPartial := &Node{ID: 4, Tuples: Relation{{9, 8}, {9, 3}, {9, 5}}, Lock: &sync.Mutex{}}
	tPartial.SetBag([]string{"V", "Z"})
	dPartial.AddChild(rPartial)
	rPartial.AddChild(sPartial)
	rPartial.AddChild(tPartial)

	//creating output
	dOutput := &Node{ID: 1, Tuples: Relation{{3, 8}, {3, 7}}}
	dOutput.SetBag([]string{"Y", "P"})
	rOutput := &Node{ID: 2, Tuples: Relation{{3, 8, 9}, {3, 8, 3}}}
	rOutput.SetBag([]string{"Y", "Z", "U"})
	sOutput := &Node{ID: 3, Tuples: Relation{{8, 3, 8}, {8, 9, 4}}}
	sOutput.SetBag([]string{"Z", "U", "W"})
	tOutput := &Node{ID: 4, Tuples: Relation{{9, 8}}}
	tOutput.SetBag([]string{"V", "Z"})
	dOutput.AddChild(rOutput)
	rOutput.AddChild(sOutput)
	rOutput.AddChild(tOutput)

	s1 := map[string]int{
		"P": 7,
		"U": 9,
		"V": 9,
		"W": 4,
		"Y": 3,
		"Z": 8,
	}
	s2 := map[string]int{
		"P": 8,
		"U": 9,
		"V": 9,
		"W": 4,
		"Y": 3,
		"Z": 8,
	}
	s3 := map[string]int{
		"P": 7,
		"U": 3,
		"V": 9,
		"W": 8,
		"Y": 3,
		"Z": 8,
	}
	s4 := map[string]int{
		"P": 8,
		"U": 3,
		"V": 9,
		"W": 8,
		"Y": 3,
		"Z": 8,
	}

	return dInput, dPartial, dOutput, []ctr.Solution{s1, s2, s3, s4}
}

func test2Data() (*Node, *Node, *Node, []ctr.Solution) {
	//creating input
	dInput := &Node{ID: 1, Tuples: Relation{{3, 8}, {3, 7}, {5, 7}, {6, 7}}, Lock: &sync.Mutex{}}
	dInput.SetBag([]string{"Y", "P"})
	rInput := &Node{ID: 2, Tuples: Relation{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 4}, {3, 8, 3}, {8, 9, 4}, {9, 4, 7}}, Lock: &sync.Mutex{}}
	rInput.SetBag([]string{"Y", "Z", "U"})
	aInput := &Node{ID: 5, Tuples: Relation{{8, 4}, {8, 7}, {4, 9}, {3, 5}}, Lock: &sync.Mutex{}}
	aInput.SetBag([]string{"P", "C"})
	sInput := &Node{ID: 3, Tuples: Relation{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 4}, {3, 8, 3}, {8, 9, 4}, {9, 4, 7}}, Lock: &sync.Mutex{}}
	sInput.SetBag([]string{"Z", "U", "W"})
	tInput := &Node{ID: 4, Tuples: Relation{{9, 8}, {9, 3}, {9, 5}}, Lock: &sync.Mutex{}}
	tInput.SetBag([]string{"V", "Z"})
	bInput := &Node{ID: 6, Tuples: Relation{{4, 1}, {3, 2}, {5, 4}}, Lock: &sync.Mutex{}}
	bInput.SetBag([]string{"C", "A"})
	dInput.AddChild(rInput)
	dInput.AddChild(aInput)
	rInput.AddChild(sInput)
	rInput.AddChild(tInput)
	aInput.AddChild(bInput)

	// creating partially reduced
	dPartial := &Node{ID: 1, Tuples: Relation{{3, 8}}, Lock: &sync.Mutex{}}
	dPartial.SetBag([]string{"Y", "P"})
	rPartial := &Node{ID: 2, Tuples: Relation{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 3}}, Lock: &sync.Mutex{}}
	rPartial.SetBag([]string{"Y", "Z", "U"})
	aPartial := &Node{ID: 5, Tuples: Relation{{8, 4}, {3, 5}}, Lock: &sync.Mutex{}}
	aPartial.SetBag([]string{"P", "C"})
	sPartial := &Node{ID: 3, Tuples: Relation{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 4}, {3, 8, 3}, {8, 9, 4}, {9, 4, 7}}, Lock: &sync.Mutex{}}
	sPartial.SetBag([]string{"Z", "U", "W"})
	tPartial := &Node{ID: 4, Tuples: Relation{{9, 8}, {9, 3}, {9, 5}}, Lock: &sync.Mutex{}}
	tPartial.SetBag([]string{"V", "Z"})
	bPartial := &Node{ID: 6, Tuples: Relation{{4, 1}, {3, 2}, {5, 4}}, Lock: &sync.Mutex{}}
	bPartial.SetBag([]string{"C", "A"})
	dPartial.AddChild(rPartial)
	dPartial.AddChild(aPartial)
	rPartial.AddChild(sPartial)
	rPartial.AddChild(tPartial)
	aPartial.AddChild(bPartial)

	//creating output
	dOutput := &Node{ID: 1, Tuples: Relation{{3, 8}}}
	dOutput.SetBag([]string{"Y", "P"})
	rOutput := &Node{ID: 2, Tuples: Relation{{3, 8, 9}, {3, 8, 3}}}
	rOutput.SetBag([]string{"Y", "Z", "U"})
	aOutput := &Node{ID: 5, Tuples: Relation{{8, 4}}}
	aOutput.SetBag([]string{"P", "C"})
	sOutput := &Node{ID: 3, Tuples: Relation{{8, 3, 8}, {8, 9, 4}}}
	sOutput.SetBag([]string{"Z", "U", "W"})
	tOutput := &Node{ID: 4, Tuples: Relation{{9, 8}}}
	tOutput.SetBag([]string{"V", "Z"})
	bOutput := &Node{ID: 6, Tuples: Relation{{4, 1}}}
	bOutput.SetBag([]string{"C", "A"})
	dOutput.AddChild(rOutput)
	dOutput.AddChild(aOutput)
	rOutput.AddChild(sOutput)
	rOutput.AddChild(tOutput)
	aOutput.AddChild(bOutput)

	s1 := map[string]int{
		"A": 1,
		"C": 4,
		"P": 8,
		"U": 3,
		"V": 9,
		"W": 8,
		"Y": 3,
		"Z": 8,
	}
	s2 := map[string]int{
		"A": 1,
		"C": 4,
		"P": 8,
		"U": 9,
		"V": 9,
		"W": 4,
		"Y": 3,
		"Z": 8,
	}

	return dInput, dPartial, dOutput, []ctr.Solution{s1, s2}
}
