package decomp

import (
	"sync"
	"testing"
)

func TestSeq1(t *testing.T) {
	input, output := test1Data()
	if !equals(YannakakisSeq(input), output) {
		t.Error()
	}
}

func TestSeq2(t *testing.T) {
	input, output := test2Data()
	if !equals(YannakakisSeq(input), output) {
		t.Error()
	}
}

func TestPar1(t *testing.T) {
	input, output := test1Data()
	if !equals(YannakakisPar(input), output) {
		t.Error()
	}
}

func TestPar2(t *testing.T) {
	input, output := test2Data()
	if !equals(YannakakisPar(input), output) {
		t.Error()
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

func test1Data() (*Node, *Node) {
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

	return dInput, dOutput
}

func test2Data() (*Node, *Node) {
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

	return dInput, dOutput
}
