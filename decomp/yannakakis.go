package decomp

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/dmlongo/callidus/csp"
	"github.com/dmlongo/callidus/db"
)

type Yannakakis interface {
	//csp.Solver

	// Solve of the problem represented by the given tree
	Solve() (csp.Solution, bool)

	// AllSolutions of the problem represent by the given tree
	AllSolutions() []csp.Solution

	// reduce a tree with upwards semijoins
	reduce(root *Node) bool
	// fullyReduce a tree with downwards semijoins (after reduce)
	fullyReduce(root *Node)
	// joinUpwards a tree to compute all solutions (after fullyReduce)
	joinUpwards(root *Node) ([]string, db.Relation)
}

func NewYannakakis(tree *Node, mode string) (Yannakakis, error) {
	switch mode {
	case "seq":
		return &seqY{tree: tree}, nil
	case "par":
		return &parY{tree: tree}, nil
	//case "ymca":
	//	return &ymca{tree: tree}, nil
	default:
		return nil, fmt.Errorf("%v yannakakis not implemented", mode)
	}
}

type seqY struct {
	tree *Node
	sol  csp.Solution
	all  []csp.Solution
}

func (y *seqY) Solve() (csp.Solution, bool) {
	if y.sol == nil {
		if y.reduce(y.tree) {
			// TODO backtrack
			// measure time diff of back with und ohne fullyReduce
			y.sol = csp.Solution{"": 0}
		} else {
			y.sol = csp.Solution{}
		}
	}
	if len(y.sol) == 0 {
		return y.sol, false
	} else {
		return y.sol, true
	}
}

func (y *seqY) AllSolutions() []csp.Solution { // TODO
	if y.all == nil {
		y.fullyReduce(y.tree)
		_, rel := y.joinUpwards(y.tree)

		fmt.Print("(Conversion from Relation to Solution... ")
		startConversion := time.Now()
		y.all = db.RelToSolutions(rel)
		fmt.Print("done in ", time.Since(startConversion), ") ")
	}
	return y.all
}

func (y *seqY) reduce(root *Node) bool {
	// bottom-up
	for _, child := range root.Children {
		if !y.reduce(child) {
			return false
		}
		db.Semijoin(root.Table, child.Table)
		if root.Table.Empty() {
			return false
		}
	}
	return true
}

func (y *seqY) fullyReduce(root *Node) {
	// top-down
	for _, child := range root.Children {
		db.Semijoin(child.Table, root.Table)
		y.fullyReduce(child)
	}
}

func (y *seqY) joinUpwards(curr *Node) ([]string, db.Relation) {
	for _, child := range curr.Children {
		childBag, childTuples := y.joinUpwards(child)
		child.SetBag(childBag)
		child.Table = childTuples

		currRel := db.Join(curr.Table, child.Table)
		curr.SetBag(currRel.Attributes())
		curr.Table = currRel
	}
	return curr.bag, curr.Table
}

type parY struct {
	tree *Node
	sol  csp.Solution
	all  []csp.Solution
}

func (y *parY) Solve() (csp.Solution, bool) {
	if y.sol == nil {
		if y.reduce(y.tree) {
			// TODO backtrack
			// measure time diff of back with und ohne fullyReduce
			y.sol = csp.Solution{"": 0}
		} else {
			y.sol = csp.Solution{}
		}
	}
	if len(y.sol) == 0 {
		return y.sol, false
	} else {
		return y.sol, true
	}
}

func (y *parY) AllSolutions() []csp.Solution { // TODO
	if y.all == nil {
		y.fullyReduce(y.tree)
		_, rel := y.joinUpwards(y.tree)

		fmt.Print("(Conversion from Relation to Solution... ")
		startConversion := time.Now()
		y.all = db.RelToSolutions(rel)
		fmt.Print("done in ", time.Since(startConversion), ") ")
	}
	return y.all
}

func (y *parY) reduce(root *Node) bool {
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

	return <-sat
}

func (y *parY) fullyReduce(root *Node) {
	var wg *sync.WaitGroup = &sync.WaitGroup{}
	for _, child := range root.Children {
		wg.Add(1)
		db.Semijoin(child.Table, root.Table)
		go func(c *Node) {
			y.fullyReduce(c)
			wg.Done()
		}(child)

	}
	wg.Wait()
}

func (y *parY) joinUpwards(curr *Node) ([]string, db.Relation) {
	var wg *sync.WaitGroup = &sync.WaitGroup{}
	for _, child := range curr.Children {
		wg.Add(1)
		go func(child *Node) {
			defer wg.Done()
			childBag, childTuples := y.joinUpwards(child)
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
