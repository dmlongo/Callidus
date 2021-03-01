package decomp

import "testing"

func Test1(t *testing.T) {
	input, output := test1Data()
	if !equals(Yannakakis(input, true, false, ""), output) {
		t.Error()
	}
}

func Test2(t *testing.T) {
	input, output := test2Data()
	if !equals(Yannakakis(input, true, false, ""), output) {
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
	if len(node1.PossibleValues) != len(node2.PossibleValues) {
		return false
	}
	for i := range node1.PossibleValues {
		if len(node1.PossibleValues[i]) != len(node2.PossibleValues[i]) {
			return false
		}
		for j := range node1.PossibleValues[i] {
			if node1.PossibleValues[i][j] != node2.PossibleValues[i][j] {
				return false
			}
		}
	}
	return true
}

func test1Data() (*Node, *Node) {
	//creating input
	dInput := &Node{ID: 1, Bag: []string{"Y", "P"}, PossibleValues: [][]int{{3, 8}, {3, 7}, {5, 7}, {6, 7}}}
	rInput := &Node{ID: 2, Bag: []string{"Y", "Z", "U"}, PossibleValues: [][]int{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 4}, {3, 8, 3}, {8, 9, 4}, {9, 4, 7}}}
	sInput := &Node{ID: 3, Bag: []string{"Z", "U", "W"}, PossibleValues: [][]int{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 4}, {3, 8, 3}, {8, 9, 4}, {9, 4, 7}}}
	tInput := &Node{ID: 4, Bag: []string{"V", "Z"}, PossibleValues: [][]int{{9, 8}, {9, 3}, {9, 5}}}
	dInput.AddChild(rInput)
	rInput.AddChild(sInput)
	rInput.AddChild(tInput)

	//creating output
	dOutput := &Node{ID: 1, Bag: []string{"Y", "P"}, PossibleValues: [][]int{{3, 8}, {3, 7}}}
	rOutput := &Node{ID: 2, Bag: []string{"Y", "Z", "U"}, PossibleValues: [][]int{{3, 8, 9}, {3, 8, 3}}}
	sOutput := &Node{ID: 3, Bag: []string{"Z", "U", "W"}, PossibleValues: [][]int{{8, 3, 8}, {8, 9, 4}}}
	tOutput := &Node{ID: 4, Bag: []string{"V", "Z"}, PossibleValues: [][]int{{9, 8}}}
	dOutput.AddChild(rOutput)
	rOutput.AddChild(sOutput)
	rOutput.AddChild(tOutput)

	return dInput, dOutput
}

func test2Data() (*Node, *Node) {
	//creating input
	dInput := &Node{ID: 1, Bag: []string{"Y", "P"}, PossibleValues: [][]int{{3, 8}, {3, 7}, {5, 7}, {6, 7}}}
	rInput := &Node{ID: 2, Bag: []string{"Y", "Z", "U"}, PossibleValues: [][]int{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 4}, {3, 8, 3}, {8, 9, 4}, {9, 4, 7}}}
	aInput := &Node{ID: 5, Bag: []string{"P", "C"}, PossibleValues: [][]int{{8, 4}, {8, 7}, {4, 9}, {3, 5}}}
	sInput := &Node{ID: 3, Bag: []string{"Z", "U", "W"}, PossibleValues: [][]int{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 4}, {3, 8, 3}, {8, 9, 4}, {9, 4, 7}}}
	tInput := &Node{ID: 4, Bag: []string{"V", "Z"}, PossibleValues: [][]int{{9, 8}, {9, 3}, {9, 5}}}
	bInput := &Node{ID: 6, Bag: []string{"C", "A"}, PossibleValues: [][]int{{4, 1}, {3, 2}, {5, 4}}}
	dInput.AddChild(rInput)
	dInput.AddChild(aInput)
	rInput.AddChild(sInput)
	rInput.AddChild(tInput)
	aInput.AddChild(bInput)

	//creating output
	dOutput := &Node{ID: 1, Bag: []string{"Y", "P"}, PossibleValues: [][]int{{3, 8}}}
	rOutput := &Node{ID: 2, Bag: []string{"Y", "Z", "U"}, PossibleValues: [][]int{{3, 8, 9}, {3, 8, 3}}}
	aOutput := &Node{ID: 5, Bag: []string{"P", "C"}, PossibleValues: [][]int{{8, 4}}}
	sOutput := &Node{ID: 3, Bag: []string{"Z", "U", "W"}, PossibleValues: [][]int{{8, 3, 8}, {8, 9, 4}}}
	tOutput := &Node{ID: 4, Bag: []string{"V", "Z"}, PossibleValues: [][]int{{9, 8}}}
	bOutput := &Node{ID: 6, Bag: []string{"C", "A"}, PossibleValues: [][]int{{4, 1}}}
	dOutput.AddChild(rOutput)
	dOutput.AddChild(aOutput)
	rOutput.AddChild(sOutput)
	rOutput.AddChild(tOutput)
	aOutput.AddChild(bOutput)

	return dInput, dOutput
}
