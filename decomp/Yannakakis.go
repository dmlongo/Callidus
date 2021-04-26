package decomp

import (
	"fmt"
	"sync"
	"time"

	"github.com/dmlongo/callidus/ctr"
)

// YannakakisSeq performs the sequential Yannakakis' algorithm
func YannakakisSeq(root *Node) (*Node, bool) {
	// bottom-up
	for _, child := range root.Children {
		if _, sat := YannakakisSeq(child); !sat {
			return nil, false
		}
		Semijoin(root.Tuples, child.Tuples)
		if root.Tuples.Empty() {
			return nil, false
		}
	}
	return root, true
}

// YannakakisPar performs the parallel Yannakakis' algorithm
func YannakakisPar(root *Node) (*Node, bool) {
	// bottom-up
	var wg *sync.WaitGroup = &sync.WaitGroup{}
	for _, child := range root.Children {
		wg.Add(1)
		go func(child *Node) (*Node, bool) { // TODO implement early termination correctly
			defer wg.Done()
			if _, sat := YannakakisPar(child); !sat {
				return nil, false
			}
			root.Lock.Lock()
			Semijoin(root.Tuples, child.Tuples)
			root.Lock.Unlock()
			if root.Tuples.Empty() {
				return nil, false
			}
			return child, true
		}(child)
	}
	wg.Wait()
	return root, true
}

// FullyReduceRelationsSeq after first bottom-up reduction sequentially
func FullyReduceRelationsSeq(root *Node) *Node {
	// top-down
	for _, child := range root.Children {
		Semijoin(child.Tuples, root.Tuples)
		FullyReduceRelationsSeq(child)
	}
	return root
}

// FullyReduceRelationsPar after first bottom-up reduction in parallel
func FullyReduceRelationsPar(root *Node) *Node {
	var wg *sync.WaitGroup = &sync.WaitGroup{}
	for _, child := range root.Children {
		wg.Add(1)
		Semijoin(child.Tuples, root.Tuples)
		go func(c *Node) {
			FullyReduceRelationsPar(c)
			wg.Done()
		}(child)

	}
	wg.Wait()
	return root
}

// ComputeAllSolutionsSeq from fully reduced relations
func ComputeAllSolutionsSeq(root *Node) []ctr.Solution {
	_, rel := computeBottomUpSeq(root)

	fmt.Print("(Conversion from Relation to Solution... ")
	startConversion := time.Now()
	allSolutions := ToSolutions(rel)
	fmt.Print("done in ", time.Since(startConversion), ") ")
	return allSolutions
}

func computeBottomUpSeq(curr *Node) ([]string, Relation) {
	for _, child := range curr.Children {
		childBag, childTuples := computeBottomUpSeq(child)
		child.SetBag(childBag)
		child.Tuples = childTuples

		currRel := Join(curr.Tuples, child.Tuples)
		curr.SetBag(currRel.Attributes())
		curr.Tuples = currRel
	}
	return curr.bag, curr.Tuples
}

// ComputeAllSolutionsPar from fully reduced relations in parallel
func ComputeAllSolutionsPar(root *Node) []ctr.Solution {
	_, rel := computeBottomUpPar(root)

	fmt.Print("(Conversion from Relation to Solution... ")
	startConversion := time.Now()
	allSolutions := ToSolutions(rel)
	fmt.Print("done in ", time.Since(startConversion), ") ")
	return allSolutions
}

func computeBottomUpPar(curr *Node) ([]string, Relation) {
	var wg *sync.WaitGroup = &sync.WaitGroup{}
	for _, child := range curr.Children {
		wg.Add(1)
		go func(child *Node) {
			defer wg.Done()
			childBag, childTuples := computeBottomUpPar(child)
			child.SetBag(childBag)
			child.Tuples = childTuples

			curr.Lock.Lock()
			currRel := Join(curr.Tuples, child.Tuples)
			curr.SetBag(currRel.Attributes())
			curr.Tuples = currRel
			curr.Lock.Unlock()
		}(child)
	}
	wg.Wait()
	return curr.bag, curr.Tuples
}
