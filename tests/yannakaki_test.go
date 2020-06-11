package tests

import (
	. "../../CSP_Project/computation"
	. "../../CSP_Project/hyperTree"
	"testing"
)

func Test1(t *testing.T) {
	input, output := test1Data()
	if !equals(ParallelYannakaki(input), output) {
		t.Error()
	}
}

func Test2(t *testing.T) {
	input, output := test2Data()
	if !equals(ParallelYannakaki(input), output) {
		t.Error()
	}
}

func equals(node1 *Node, node2 *Node) bool {
	if !node1.SamePossibleValues(node2) {
		return false
	}
	for i := range node1.Sons {
		if !equals(node1.Sons[i], node2.Sons[i]) {
			return false
		}
	}
	return true
}

func test1Data() (*Node, *Node) {
	//creating input
	dInput := &Node{Id: 1, Variables: []string{"Y", "P"}, PossibleValues: [][]int{{3, 8}, {3, 7}, {5, 7}, {6, 7}}}
	rInput := &Node{Id: 2, Variables: []string{"Y", "Z", "U"}, PossibleValues: [][]int{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 4}, {3, 8, 3}, {8, 9, 4}, {9, 4, 7}}}
	sInput := &Node{Id: 3, Variables: []string{"Z", "U", "W"}, PossibleValues: [][]int{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 4}, {3, 8, 3}, {8, 9, 4}, {9, 4, 7}}}
	tInput := &Node{Id: 4, Variables: []string{"V", "Z"}, PossibleValues: [][]int{{9, 8}, {9, 3}, {9, 5}}}
	dInput.AddSon(rInput)
	rInput.AddSon(sInput)
	rInput.AddSon(tInput)

	//creating output
	dOutput := &Node{Id: 1, Variables: []string{"Y", "P"}, PossibleValues: [][]int{{3, 8}, {3, 7}}}
	rOutput := &Node{Id: 2, Variables: []string{"Y", "Z", "U"}, PossibleValues: [][]int{{3, 8, 9}, {3, 8, 3}}}
	sOutput := &Node{Id: 3, Variables: []string{"Z", "U", "W"}, PossibleValues: [][]int{{8, 3, 8}, {8, 9, 4}}}
	tOutput := &Node{Id: 4, Variables: []string{"V", "Z"}, PossibleValues: [][]int{{9, 8}}}
	dOutput.AddSon(rOutput)
	rOutput.AddSon(sOutput)
	rOutput.AddSon(tOutput)

	return dInput, dOutput
}

func test2Data() (*Node, *Node) {
	//creating input
	dInput := &Node{Id: 1, Variables: []string{"Y", "P"}, PossibleValues: [][]int{{3, 8}, {3, 7}, {5, 7}, {6, 7}}}
	rInput := &Node{Id: 2, Variables: []string{"Y", "Z", "U"}, PossibleValues: [][]int{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 4}, {3, 8, 3}, {8, 9, 4}, {9, 4, 7}}}
	aInput := &Node{Id: 5, Variables: []string{"P", "C"}, PossibleValues: [][]int{{8, 4}, {8, 7}, {4, 9}, {3, 5}}}
	sInput := &Node{Id: 3, Variables: []string{"Z", "U", "W"}, PossibleValues: [][]int{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 4}, {3, 8, 3}, {8, 9, 4}, {9, 4, 7}}}
	tInput := &Node{Id: 4, Variables: []string{"V", "Z"}, PossibleValues: [][]int{{9, 8}, {9, 3}, {9, 5}}}
	bInput := &Node{Id: 6, Variables: []string{"C", "A"}, PossibleValues: [][]int{{4, 1}, {3, 2}, {5, 4}}}
	dInput.AddSon(rInput)
	dInput.AddSon(aInput)
	rInput.AddSon(sInput)
	rInput.AddSon(tInput)
	aInput.AddSon(bInput)

	//creating output
	dOutput := &Node{Id: 1, Variables: []string{"Y", "P"}, PossibleValues: [][]int{{3, 8}}}
	rOutput := &Node{Id: 2, Variables: []string{"Y", "Z", "U"}, PossibleValues: [][]int{{3, 8, 9}, {3, 8, 3}}}
	aOutput := &Node{Id: 5, Variables: []string{"P", "C"}, PossibleValues: [][]int{{8, 4}}}
	sOutput := &Node{Id: 3, Variables: []string{"Z", "U", "W"}, PossibleValues: [][]int{{8, 3, 8}, {8, 9, 4}}}
	tOutput := &Node{Id: 4, Variables: []string{"V", "Z"}, PossibleValues: [][]int{{9, 8}}}
	bOutput := &Node{Id: 6, Variables: []string{"C", "A"}, PossibleValues: [][]int{{4, 1}}}
	dOutput.AddSon(rOutput)
	dOutput.AddSon(aOutput)
	rOutput.AddSon(sOutput)
	rOutput.AddSon(tOutput)
	aOutput.AddSon(bOutput)

	return dInput, dOutput
}
