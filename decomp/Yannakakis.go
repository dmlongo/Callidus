package decomp

import (
	"fmt"
	"sync"
	"time"

	"github.com/dmlongo/callidus/ctr"
)

/*
// IDJoiningIndex is.. I still don't know
type IDJoiningIndex struct {
	x int
	y int
}

// MyMap is a map with a lock
type MyMap struct {
	hash map[IDJoiningIndex][][]int
	lock *sync.RWMutex
}

var joiningIndex *MyMap

func init() {
	joiningIndex = &MyMap{}
	joiningIndex.hash = make(map[IDJoiningIndex][][]int)
	joiningIndex.lock = &sync.RWMutex{}
}
*/

// YannakakisSeq performs the sequential Yannakakis' algorithm
func YannakakisSeq(root *Node) *Node {
	if len(root.Children) != 0 {
		bottomUpSeq(root)
		topDownSeq(root)
	}
	return root
}

func bottomUpSeq(curr *Node) {
	for _, child := range curr.Children {
		bottomUpSeq(child)
	}
	if curr.Parent != nil {
		semiJoin(curr.Parent, curr)
	}
}

func topDownSeq(curr *Node) {
	for _, child := range curr.Children {
		semiJoin(child, curr)
		topDownSeq(child)
	}
}

// YannakakisPar performs the parrallel Yannakakis' algorithm
func YannakakisPar(root *Node) *Node {
	if len(root.Children) != 0 {
		bottomUpPar(root)
		topDownPar(root)
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
		semiJoin(curr.Parent, curr)
		curr.Parent.Lock.Unlock()
	}
}

func topDownPar(curr *Node) {
	var wg *sync.WaitGroup = &sync.WaitGroup{}
	wg.Add(len(curr.Children))
	for _, child := range curr.Children {
		semiJoin(child, curr)
		go func(c *Node) {
			topDownPar(c)
			wg.Done()
		}(child)

	}
	wg.Wait()
}

func semiJoin(left *Node, right *Node) {
	joinIdx := findJoinIndices(left, right)

	var tupToDel []int
	for i, leftTup := range left.Tuples {
		delete := true
		for _, rightTup := range right.Tuples {
			if match(leftTup, rightTup, joinIdx) {
				delete = false
				break
			}
		}
		if delete {
			tupToDel = append(tupToDel, i)
		}
	}

	update(left, tupToDel)
}

func match(left []int, right []int, joinIndex [][]int) bool {
	for _, z := range joinIndex {
		if left[z[0]] != right[z[1]] {
			return false
		}
	}
	return true
}

func update(n *Node, toDel []int) { // TODO shouldn't this method be synchronized?
	if len(toDel) > 0 {
		newSize := len(n.Tuples) - len(toDel)
		newTuples := make(Relation, 0, newSize)
		if newSize > 0 { // TODO what does newSize == 0 mean? unsat?
			i := 0
			for _, j := range toDel {
				newTuples = append(newTuples, n.Tuples[i:j]...)
				i = j + 1
			}
			newTuples = append(newTuples, n.Tuples[i:]...)
		}
		n.Tuples = newTuples
	}
}

func findJoinIndices(left *Node, right *Node) [][]int {
	var out [][]int
	for iLeft, varLeft := range left.Bag() {
		if iRight := right.Position(varLeft); iRight >= 0 {
			out = append(out, []int{iLeft, iRight})
		}
	}
	return out
}

/*
func searchJoiningIndex(left *Node, right *Node) [][]int {
	var joinIndices [][]int
	joiningIndex.lock.RLock()
	val, ok := joiningIndex.hash[IDJoiningIndex{x: left.ID, y: right.ID}]
	joiningIndex.lock.RUnlock()
	if ok {
		joinIndices = val
	} else {
		var invJoinIndices [][]int
		for iLeft, varLeft := range left.Bag() {
			if iRight := right.Position(varLeft); iRight >= 0 {
				joinIndices = append(joinIndices, []int{iLeft, iRight})
				invJoinIndices = append(invJoinIndices, []int{iRight, iLeft})
			}
		}
		joiningIndex.lock.Lock()
		joiningIndex.hash[IDJoiningIndex{x: right.ID, y: left.ID}] = invJoinIndices
		joiningIndex.lock.Unlock()
	}
	return joinIndices
}
*/

// ComputeAllSolutions from fully reduced relations
func ComputeAllSolutions(root *Node) []ctr.Solution {
	vars, rel := computeBottomUp(root)

	fmt.Print("(Conversion from Relation to Solution... ")
	startConversion := time.Now()
	var allSolutions []ctr.Solution
	for _, tup := range rel {
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

		currBag, currTuples := join(curr, child)
		curr.SetBag(currBag)
		curr.Tuples = currTuples
	}
	return curr.bag, curr.Tuples
}

func join(left *Node, right *Node) (outVars []string, outRel Relation) {
	outVars = newBag(left.bag, left.bagSet, right.bag)
	joinIdx := findJoinIndices(left, right)
	for _, lTup := range left.Tuples {
		for _, rTup := range right.Tuples {
			if match(lTup, rTup, joinIdx) {
				outTup := newTuple(outVars, lTup, rTup, right.bagSet)
				outRel = append(outRel, outTup)
			}
		}
	}

	return
}

func newBag(bagL []string, bagSetL map[string]int, bagR []string) []string {
	var out []string
	out = append(out, bagL...)
	for _, v := range bagR {
		if _, ok := bagSetL[v]; !ok {
			out = append(out, v)
		}
	}
	return out
}

func newTuple(vars []string, lTup []int, rTup []int, rBagSet map[string]int) []int {
	out := make([]int, 0, len(vars))
	out = append(out, lTup...)
	for _, v := range vars[len(lTup):] {
		i := rBagSet[v]
		out = append(out, rTup[i])
	}
	return out
}
