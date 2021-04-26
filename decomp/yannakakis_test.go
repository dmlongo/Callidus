package decomp

import (
	"sync"
	"testing"

	"github.com/dmlongo/callidus/ctr"
)

func TestSeq1(t *testing.T) {
	input, partial, output, sols := test1Data()
	if rel, sat := YannakakisSeq(input); !sat || !equals(rel, partial) {
		if !sat {
			t.Error("y(input) is unsat!")
		}
		t.Error("y(input) != partial")
	}
	if !equals(FullyReduceRelationsSeq(input), output) {
		t.Error("y(partial) != output")
	}
	if !solEquals(ComputeAllSolutionsSeq(input), sols) {
		t.Error("y(output) != solutions")
	}
}

func TestSeq2(t *testing.T) {
	input, partial, output, sols := test2Data()
	if rel, sat := YannakakisSeq(input); !sat || !equals(rel, partial) {
		if !sat {
			t.Error("y(input) is unsat!")
		}
		t.Error("y(input) != partial")
	}
	if !equals(FullyReduceRelationsSeq(input), output) {
		t.Error("y(partial) != output")
	}
	if !solEquals(ComputeAllSolutionsSeq(input), sols) {
		t.Error("y(output) != solutions")
	}
}

func TestPar1(t *testing.T) {
	input, partial, output, sols := test1Data()
	if rel, sat := YannakakisPar(input); !sat || !equals(rel, partial) {
		if !sat {
			t.Error("y(input) is unsat!")
		}
		t.Error("y(input) != partial")
	}
	if !equals(FullyReduceRelationsPar(input), output) {
		t.Error("y(partial) != output")
	}
	if !solEquals(ComputeAllSolutionsPar(input), sols) {
		t.Error("y(output) != solutions")
	}
}

func TestPar2(t *testing.T) {
	input, partial, output, sols := test2Data()
	if rel, sat := YannakakisPar(input); !sat || !equals(rel, partial) {
		if !sat {
			t.Error("y(input) is unsat!")
		}
		t.Error("y(input) != partial")
	}
	if !equals(FullyReduceRelationsPar(input), output) {
		t.Error("y(partial) != output")
	}
	if !solEquals(ComputeAllSolutionsPar(input), sols) {
		t.Error("y(output) != solutions")
	}
}

func TestYMCA1(t *testing.T) {
	input, partial, output, _ := test1Data()
	sat, err := YMCA(input)
	if err != nil {
		panic(err)
	}
	if !sat {
		t.Error("y(input) is unsat, expected sat")
	}
	if !equals(input, partial) {
		t.Error("y(input) != partial")
	}

	err = YMCAFullReduce()
	if err != nil {
		panic(err)
	}
	if !equals(input, output) {
		t.Error("y(partial) != output")
	}
	//if !equals(FullyReduceRelationsPar(partial), output) {
	//	t.Error("y(partial) != output")
	//}
	//if !solEquals(ComputeAllSolutions(output), sols) {
	//	t.Error("y(output) != solutions")
	//}
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
	if len(node1.Tuples.Tuples()) != len(node2.Tuples.Tuples()) {
		return false
	}
	for i := range node1.Tuples.Tuples() {
		if len(node1.Tuples.Tuples()[i]) != len(node2.Tuples.Tuples()[i]) {
			return false
		}
		for j := range node1.Tuples.Tuples()[i] {
			if node1.Tuples.Tuples()[i][j] != node2.Tuples.Tuples()[i][j] {
				return false
			}
		}
	}
	return true
}

func test1Data() (*Node, *Node, *Node, []ctr.Solution) {
	//creating input
	dAttrs := []string{"Y", "P"}
	dRel := []Tuple{{3, 8}, {3, 7}, {5, 7}, {6, 7}}
	dInput := &Node{ID: 1, Tuples: InitializedRelation(dAttrs, dRel), Lock: &sync.Mutex{}}
	dInput.SetBag(dAttrs)

	rAttrs := []string{"Y", "Z", "U"}
	rRel := []Tuple{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 4}, {3, 8, 3}, {8, 9, 4}, {9, 4, 7}}
	rInput := &Node{ID: 2, Tuples: InitializedRelation(rAttrs, rRel), Lock: &sync.Mutex{}}
	rInput.SetBag(rAttrs)

	sAttrs := []string{"Z", "U", "W"}
	sRel := []Tuple{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 4}, {3, 8, 3}, {8, 9, 4}, {9, 4, 7}}
	sInput := &Node{ID: 3, Tuples: InitializedRelation(sAttrs, sRel), Lock: &sync.Mutex{}}
	sInput.SetBag(sAttrs)

	tAttrs := []string{"V", "Z"}
	tRel := []Tuple{{9, 8}, {9, 3}, {9, 5}}
	tInput := &Node{ID: 4, Tuples: InitializedRelation(tAttrs, tRel), Lock: &sync.Mutex{}}
	tInput.SetBag(tAttrs)

	dInput.AddChild(rInput)
	rInput.AddChild(sInput)
	rInput.AddChild(tInput)

	// creating partially reduced
	dPartRel := []Tuple{{3, 8}, {3, 7}}
	dPartial := &Node{ID: 1, Tuples: InitializedRelation(dAttrs, dPartRel), Lock: &sync.Mutex{}}
	dPartial.SetBag(dAttrs)

	rPartRel := []Tuple{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 3}}
	rPartial := &Node{ID: 2, Tuples: InitializedRelation(rAttrs, rPartRel), Lock: &sync.Mutex{}}
	rPartial.SetBag(rAttrs)

	sPartRel := []Tuple{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 4}, {3, 8, 3}, {8, 9, 4}, {9, 4, 7}}
	sPartial := &Node{ID: 3, Tuples: InitializedRelation(sAttrs, sPartRel), Lock: &sync.Mutex{}}
	sPartial.SetBag(sAttrs)

	tPartRel := []Tuple{{9, 8}, {9, 3}, {9, 5}}
	tPartial := &Node{ID: 4, Tuples: InitializedRelation(tAttrs, tPartRel), Lock: &sync.Mutex{}}
	tPartial.SetBag(tAttrs)

	dPartial.AddChild(rPartial)
	rPartial.AddChild(sPartial)
	rPartial.AddChild(tPartial)

	//creating output
	dOutRel := []Tuple{{3, 8}, {3, 7}}
	dOutput := &Node{ID: 1, Tuples: InitializedRelation(dAttrs, dOutRel), Lock: &sync.Mutex{}}
	dOutput.SetBag(dAttrs)

	rOutRel := []Tuple{{3, 8, 9}, {3, 8, 3}}
	rOutput := &Node{ID: 2, Tuples: InitializedRelation(rAttrs, rOutRel), Lock: &sync.Mutex{}}
	rOutput.SetBag(rAttrs)

	sOutRel := []Tuple{{8, 3, 8}, {8, 9, 4}}
	sOutput := &Node{ID: 3, Tuples: InitializedRelation(sAttrs, sOutRel), Lock: &sync.Mutex{}}
	sOutput.SetBag(sAttrs)

	tOutRel := []Tuple{{9, 8}}
	tOutput := &Node{ID: 4, Tuples: InitializedRelation(tAttrs, tOutRel), Lock: &sync.Mutex{}}
	tOutput.SetBag(tAttrs)

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
	dAttrs := []string{"Y", "P"}
	dRel := []Tuple{{3, 8}, {3, 7}, {5, 7}, {6, 7}}
	dInput := &Node{ID: 1, Tuples: InitializedRelation(dAttrs, dRel), Lock: &sync.Mutex{}}
	dInput.SetBag(dAttrs)

	rAttrs := []string{"Y", "Z", "U"}
	rRel := []Tuple{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 4}, {3, 8, 3}, {8, 9, 4}, {9, 4, 7}}
	rInput := &Node{ID: 2, Tuples: InitializedRelation(rAttrs, rRel), Lock: &sync.Mutex{}}
	rInput.SetBag(rAttrs)

	aAttrs := []string{"P", "C"}
	aRel := []Tuple{{8, 4}, {8, 7}, {4, 9}, {3, 5}}
	aInput := &Node{ID: 5, Tuples: InitializedRelation(aAttrs, aRel), Lock: &sync.Mutex{}}
	aInput.SetBag(aAttrs)

	sAttrs := []string{"Z", "U", "W"}
	sRel := []Tuple{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 4}, {3, 8, 3}, {8, 9, 4}, {9, 4, 7}}
	sInput := &Node{ID: 3, Tuples: InitializedRelation(sAttrs, sRel), Lock: &sync.Mutex{}}
	sInput.SetBag(sAttrs)

	tAttrs := []string{"V", "Z"}
	tRel := []Tuple{{9, 8}, {9, 3}, {9, 5}}
	tInput := &Node{ID: 4, Tuples: InitializedRelation(tAttrs, tRel), Lock: &sync.Mutex{}}
	tInput.SetBag(tAttrs)

	bAttrs := []string{"C", "A"}
	bRel := []Tuple{{4, 1}, {3, 2}, {5, 4}}
	bInput := &Node{ID: 6, Tuples: InitializedRelation(bAttrs, bRel), Lock: &sync.Mutex{}}
	bInput.SetBag(bAttrs)

	dInput.AddChild(rInput)
	dInput.AddChild(aInput)
	rInput.AddChild(sInput)
	rInput.AddChild(tInput)
	aInput.AddChild(bInput)

	// creating partially reduced
	dPartRel := []Tuple{{3, 8}}
	dPartial := &Node{ID: 1, Tuples: InitializedRelation(dAttrs, dPartRel), Lock: &sync.Mutex{}}
	dPartial.SetBag(dAttrs)

	rPartRel := []Tuple{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 3}}
	rPartial := &Node{ID: 2, Tuples: InitializedRelation(rAttrs, rPartRel), Lock: &sync.Mutex{}}
	rPartial.SetBag(rAttrs)

	aPartRel := []Tuple{{8, 4}, {3, 5}}
	aPartial := &Node{ID: 5, Tuples: InitializedRelation(aAttrs, aPartRel), Lock: &sync.Mutex{}}
	aPartial.SetBag(aAttrs)

	sPartRel := []Tuple{{3, 8, 9}, {9, 3, 8}, {8, 3, 8}, {3, 8, 4}, {3, 8, 3}, {8, 9, 4}, {9, 4, 7}}
	sPartial := &Node{ID: 3, Tuples: InitializedRelation(sAttrs, sPartRel), Lock: &sync.Mutex{}}
	sPartial.SetBag(sAttrs)

	tPartRel := []Tuple{{9, 8}, {9, 3}, {9, 5}}
	tPartial := &Node{ID: 4, Tuples: InitializedRelation(tAttrs, tPartRel), Lock: &sync.Mutex{}}
	tPartial.SetBag(tAttrs)

	bPartRel := []Tuple{{4, 1}, {3, 2}, {5, 4}}
	bPartial := &Node{ID: 6, Tuples: InitializedRelation(bAttrs, bPartRel), Lock: &sync.Mutex{}}
	bPartial.SetBag(bAttrs)

	dPartial.AddChild(rPartial)
	dPartial.AddChild(aPartial)
	rPartial.AddChild(sPartial)
	rPartial.AddChild(tPartial)
	aPartial.AddChild(bPartial)

	//creating output
	dOutRel := []Tuple{{3, 8}}
	dOutput := &Node{ID: 1, Tuples: InitializedRelation(dAttrs, dOutRel), Lock: &sync.Mutex{}}
	dOutput.SetBag(dAttrs)

	rOutRel := []Tuple{{3, 8, 9}, {3, 8, 3}}
	rOutput := &Node{ID: 2, Tuples: InitializedRelation(rAttrs, rOutRel), Lock: &sync.Mutex{}}
	rOutput.SetBag(rAttrs)

	aOutRel := []Tuple{{8, 4}}
	aOutput := &Node{ID: 5, Tuples: InitializedRelation(aAttrs, aOutRel), Lock: &sync.Mutex{}}
	aOutput.SetBag(aAttrs)

	sOutRel := []Tuple{{8, 3, 8}, {8, 9, 4}}
	sOutput := &Node{ID: 3, Tuples: InitializedRelation(sAttrs, sOutRel), Lock: &sync.Mutex{}}
	sOutput.SetBag(sAttrs)

	tOutRel := []Tuple{{9, 8}}
	tOutput := &Node{ID: 4, Tuples: InitializedRelation(tAttrs, tOutRel), Lock: &sync.Mutex{}}
	tOutput.SetBag(tAttrs)

	bOutRel := []Tuple{{4, 1}}
	bOutput := &Node{ID: 6, Tuples: InitializedRelation(bAttrs, bOutRel), Lock: &sync.Mutex{}}
	bOutput.SetBag(bAttrs)

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
