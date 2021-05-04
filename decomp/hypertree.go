package decomp

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/dmlongo/callidus/db"
)

// TODO make it really a tree!

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

	Table db.Relation
	Lock  *sync.Mutex
}

func NewNode(id int, vars []string, edges []string) *Node {
	n := Node{ID: id}
	n.SetBag(vars)
	n.SetCover(edges)
	n.Table = db.NewRelation(vars)
	n.Lock = &sync.Mutex{}
	return &n
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
			m := NewNode(*maxID, e.vertices, []string{e.name})
			n.AddChild(m)
			*tree = append(*tree, m)
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

func PrintTreeRelations(n *Node) {
	fmt.Println("Rel-" + strconv.Itoa(n.ID))
	fmt.Println(db.RelToString(n.Table))
	fmt.Println()
	for _, c := range n.Children {
		PrintTreeRelations(c)
	}
}

func Bfs(root *Node) Hypertree {
	nodes := make(Hypertree, 0)
	toVisit := make(Hypertree, 0)
	toVisit = append(toVisit, root)
	var curr *Node
	for len(toVisit) > 0 {
		curr, toVisit = toVisit[0], toVisit[1:]
		nodes = append(nodes, curr)
		toVisit = append(toVisit, curr.Children...)
	}
	return nodes
}
