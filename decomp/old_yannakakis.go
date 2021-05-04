package decomp

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/dmlongo/callidus/csp"
	"github.com/dmlongo/callidus/db"
)

// YannakakisSeq performs the sequential Yannakakis' algorithm
func YannakakisSeq(root *Node) (*Node, bool) {
	// bottom-up
	for _, child := range root.Children {
		if _, sat := YannakakisSeq(child); !sat {
			return nil, false
		}
		db.Semijoin(root.Table, child.Table)
		if root.Table.Empty() {
			return nil, false
		}
	}
	return root, true
}

// YannakakisPar performs the parallel Yannakakis' algorithm
/*func YannakakisPar(root *Node) (*Node, bool) {
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
}*/

// YannakakisPar performs the parallel Yannakakis' algorithm
func YannakakisPar(root *Node) (*Node, bool) {
	nodes := Bfs(root)
	leaves := 0

	type job struct {
		id    int
		lock  *sync.Mutex
		left  db.Relation
		right db.Relation
	}
	type result struct {
		id  int
		sat bool
	}

	jobs := make(chan *job, len(nodes))
	results := make(chan *result)
	sat := make(chan bool)
	numLeaves := make(chan int)
	quit := make(chan bool)
	defer close(quit)
	go func() {
		defer close(jobs)
		defer close(sat)
		deps := make(map[int]int)
		id2node := make(map[int]*Node)
		for _, curr := range nodes {
			numChildren := len(curr.Children)
			if numChildren == 0 && curr.Parent != nil {
				leaves++
				jobs <- &job{
					id:    curr.Parent.ID,
					lock:  curr.Parent.Lock,
					left:  curr.Parent.Table,
					right: curr.Table,
				}
			} else {
				deps[curr.ID] = numChildren
				id2node[curr.ID] = curr
			}
		}
		numLeaves <- leaves
		close(numLeaves)

		for len(deps) > 0 {
			res := <-results
			if res.sat {
				parentID := res.id
				parent := id2node[parentID]
				deps[parentID] -= 1
				if deps[parentID] == 0 {
					delete(deps, parentID)
					delete(id2node, parentID)
					if parent.Parent != nil {
						jobs <- &job{
							id:    parent.Parent.ID,
							lock:  parent.Parent.Lock,
							left:  parent.Parent.Table,
							right: parent.Table,
						}
					}
				}
			} else {
				sat <- false
				return
			}
		}
		sat <- true
	}()

	numNodes := <-numLeaves
	numWorkers := runtime.NumCPU()
	if numNodes < numWorkers {
		numWorkers = numNodes
	}
	for i := 0; i < numWorkers; i++ {
		go func() { // launch a worker
			for job := range jobs {
				job.lock.Lock()
				db.Semijoin(job.left, job.right)
				sat := !job.left.Empty()
				job.lock.Unlock()
				select {
				case results <- &result{
					id:  job.id,
					sat: sat,
				}:
				case <-quit:
					return
				}
			}
		}()
	}

	if !<-sat {
		return nil, false
	}
	return root, true
}

// FullyReduceRelationsSeq after first bottom-up reduction sequentially
func FullyReduceRelationsSeq(root *Node) *Node {
	// top-down
	for _, child := range root.Children {
		db.Semijoin(child.Table, root.Table)
		FullyReduceRelationsSeq(child)
	}
	return root
}

// FullyReduceRelationsPar after first bottom-up reduction in parallel
func FullyReduceRelationsPar(root *Node) *Node {
	var wg *sync.WaitGroup = &sync.WaitGroup{}
	for _, child := range root.Children {
		wg.Add(1)
		db.Semijoin(child.Table, root.Table)
		go func(c *Node) {
			FullyReduceRelationsPar(c)
			wg.Done()
		}(child)

	}
	wg.Wait()
	return root
}

// ComputeAllSolutionsSeq from fully reduced relations
func ComputeAllSolutionsSeq(root *Node) []csp.Solution {
	_, rel := computeBottomUpSeq(root)

	fmt.Print("(Conversion from Relation to Solution... ")
	startConversion := time.Now()
	allSolutions := db.RelToSolutions(rel)
	fmt.Print("done in ", time.Since(startConversion), ") ")
	return allSolutions
}

func computeBottomUpSeq(curr *Node) ([]string, db.Relation) {
	for _, child := range curr.Children {
		childBag, childTuples := computeBottomUpSeq(child)
		child.SetBag(childBag)
		child.Table = childTuples

		currRel := db.Join(curr.Table, child.Table)
		curr.SetBag(currRel.Attributes())
		curr.Table = currRel
	}
	return curr.bag, curr.Table
}

// ComputeAllSolutionsPar from fully reduced relations in parallel
func ComputeAllSolutionsPar(root *Node) []csp.Solution {
	_, rel := computeBottomUpPar(root)

	fmt.Print("(Conversion from Relation to Solution... ")
	startConversion := time.Now()
	allSolutions := db.RelToSolutions(rel)
	fmt.Print("done in ", time.Since(startConversion), ") ")
	return allSolutions
}

func computeBottomUpPar(curr *Node) ([]string, db.Relation) {
	var wg *sync.WaitGroup = &sync.WaitGroup{}
	for _, child := range curr.Children {
		wg.Add(1)
		go func(child *Node) {
			defer wg.Done()
			childBag, childTuples := computeBottomUpPar(child)
			child.SetBag(childBag)
			child.Table = childTuples

			curr.Lock.Lock()
			currRel := db.Join(curr.Table, child.Table)
			curr.SetBag(currRel.Attributes())
			curr.Table = currRel
			curr.Lock.Unlock()

			child.Table = nil
		}(child)
	}
	wg.Wait()
	return curr.bag, curr.Table
}
