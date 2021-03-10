package decomp

import (
	"fmt"
	"sync"
)

// TODO make it really a tree!

// Relation represent a collection of multiple tuples
type Relation [][]int

type Hypertree []*Node

// Node of a hypertree decomposition
type Node struct {
	ID       int
	Father   *Node
	Children []*Node
	Bag      []string
	Cover    []string
	Tuples   Relation

	bagSet   map[string]bool
	coverSet map[string]bool

	Lock *sync.Mutex
}

// AddChild to this node
func (n *Node) AddChild(child *Node) {
	n.Children = append(n.Children, child)
	child.Father = n
}

func (n *Node) SetBag(bag []string) { // TODO reset map
	n.Bag = bag
	n.bagSet = make(map[string]bool)
	for _, v := range bag {
		n.bagSet[v] = true
	}
}

func (n *Node) SetCover(cover []string) { // TODO reset map
	n.Cover = cover
	n.coverSet = make(map[string]bool)
	for _, e := range cover {
		n.coverSet[e] = true
	}
}

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
		for _, e := range n.Cover {
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
			m := Node{ID: *maxID, Father: n}
			m.SetBag(e.vertices)
			m.SetCover([]string{e.name})
			*tree = append(*tree, &m)
			break
		}
	}
	if err {
		panic(fmt.Sprint("Could not find a place for e=", e.vertices))
	}
}

func subset(s []string, p map[string]bool) bool {
	for _, e := range s {
		if _, ok := p[e]; !ok {
			return false
		}
	}
	return true
}
