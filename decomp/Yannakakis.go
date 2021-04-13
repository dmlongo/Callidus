package decomp

import (
	"fmt"
	"sync"
	"time"

	"github.com/dmlongo/callidus/ctr"
)

// YannakakisSeq performs the sequential Yannakakis' algorithm
func YannakakisSeq(root *Node) (*Node, bool) {
	if len(root.Children) != 0 {
		return bottomUpSeq(root)
	}
	return root, true
}

func bottomUpSeq(curr *Node) (*Node, bool) {
	for _, child := range curr.Children {
		if _, sat := bottomUpSeq(child); !sat {
			return nil, false
		}
	}
	if curr.Parent != nil {
		Semijoin(curr.Parent.Tuples, curr.Tuples)
		if curr.Parent.Tuples.Empty() {
			return nil, false
		}
	}
	return curr, true
}

// YannakakisPar performs the parrallel Yannakakis' algorithm
func YannakakisPar(root *Node) *Node {
	if len(root.Children) != 0 {
		bottomUpPar(root)
	}
	return root
}

func bottomUpPar(curr *Node) {
	var wg *sync.WaitGroup = &sync.WaitGroup{}
	for _, child := range curr.Children {
		if len(child.Children) != 0 && len(curr.Children) > 1 {
			wg.Add(1)
			go func(c *Node) {
				bottomUpPar(c)
				wg.Done()
			}(child)
		} else {
			bottomUpPar(child)
		}
	}
	wg.Wait()
	if curr.Parent != nil {
		curr.Parent.Lock.Lock()
		Semijoin(curr.Parent.Tuples, curr.Tuples)
		curr.Parent.Lock.Unlock()
	}
}

// FullyReduceRelationsSeq after first bottom-up reduction sequentially
func FullyReduceRelationsSeq(root *Node) *Node {
	if len(root.Children) != 0 {
		topDownSeq(root)
	}
	return root
}

func topDownSeq(curr *Node) {
	for _, child := range curr.Children {
		Semijoin(child.Tuples, curr.Tuples)
		topDownSeq(child)
	}
}

// FullyReduceRelationsPar after first bottom-up reduction in parallel
func FullyReduceRelationsPar(root *Node) *Node {
	if len(root.Children) != 0 {
		topDownPar(root)
	}
	return root
}

func topDownPar(curr *Node) {
	var wg *sync.WaitGroup = &sync.WaitGroup{}
	wg.Add(len(curr.Children))
	for _, child := range curr.Children {
		Semijoin(child.Tuples, curr.Tuples)
		go func(c *Node) {
			topDownPar(c)
			wg.Done()
		}(child)

	}
	wg.Wait()
}

// ComputeAllSolutions from fully reduced relations
func ComputeAllSolutions(root *Node) []ctr.Solution {
	vars, rel := computeBottomUp(root)

	fmt.Print("(Conversion from Relation to Solution... ")
	startConversion := time.Now()
	var allSolutions []ctr.Solution
	for _, tup := range rel.Tuples() {
		sol := make(ctr.Solution)
		for i, v := range vars {
			sol[v] = tup[i]
		}
		allSolutions = append(allSolutions, sol)
	}
	fmt.Print("done in ", time.Since(startConversion), ") ")
	return allSolutions
}

func computeBottomUp(curr *Node) ([]string, Relation) {
	for _, child := range curr.Children {
		childBag, childTuples := computeBottomUp(child)
		child.SetBag(childBag)
		child.Tuples = childTuples

		currRel := Join(curr.Tuples, child.Tuples)
		curr.SetBag(currRel.Attributes())
		curr.Tuples = currRel
	}
	return curr.bag, curr.Tuples
}
