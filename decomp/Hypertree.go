package decomp

import (
	"fmt"
	"sync"
)

// TODO make it really a tree!

// Relation represent a collection of multiple tuples
type Relation [][]int

// Hypertree is a slice of nodes in the tree
type Hypertree []*Node

// Node of a hypertree decomposition
type Node struct {
	ID       int
	Parent   *Node
	Children []*Node

	bag      []string
	bagSet   map[string]int
	cover    []string
	coverSet map[string]bool

	Tuples Relation
	Lock   *sync.Mutex
}

// AddChild to this node
func (n *Node) AddChild(child *Node) {
	n.Children = append(n.Children, child)
	child.Parent = n
}

// SetBag inits the bag of a node
func (n *Node) SetBag(bag []string) {
	n.bag = bag
	n.bagSet = make(map[string]int)
	for i, v := range bag {
		n.bagSet[v] = i
	}
}

// Bag of a node
func (n *Node) Bag() []string {
	return n.bag
}

// Position of the variable v in the bag of a node
func (n *Node) Position(v string) int {
	if p, ok := n.bagSet[v]; ok {
		return p
	}
	return -1
}

// SetCover inits the edge cover of a node
func (n *Node) SetCover(cover []string) {
	n.cover = cover
	n.coverSet = make(map[string]bool)
	for _, e := range cover {
		n.coverSet[e] = true
	}
}

// Cover of a node
func (n *Node) Cover() []string {
	return n.cover
}

// Complete a hypertree wrt a hypergraph
func (tree *Hypertree) Complete(hg Hypergraph) {
	labels, maxID := tree.coveredEdges()
	for k, e := range hg {
		if _, ok := labels[k]; !ok {
			tree.attach(e, &maxID)
		}
	}
}

func (tree *Hypertree) coveredEdges() (map[string]bool, int) {
	res := make(map[string]bool)
	maxID := 0
	for _, n := range *tree {
		for _, e := range n.cover {
			res[e] = true
		}
		if n.ID > maxID {
			maxID = n.ID
		}
	}
	return res, maxID
}

func (tree *Hypertree) attach(e Edge, maxID *int) {
	err := true
	for _, n := range *tree {
		if subset(e.vertices, n.bagSet) {
			err = false
			*maxID = *maxID + 1
			m := Node{ID: *maxID, Parent: n}
			m.SetBag(e.vertices)
			m.SetCover([]string{e.name})
			n.AddChild(&m)
			*tree = append(*tree, &m)
			break
		}
	}
	if err {
		panic(fmt.Sprint("Could not find a place for e=", e.vertices))
	}
}

func subset(s []string, p map[string]int) bool {
	for _, e := range s {
		if _, ok := p[e]; !ok {
			return false
		}
	}
	return true
}
