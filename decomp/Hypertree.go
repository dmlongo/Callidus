package decomp

import "sync"

// TODO make it really a tree!

// Node of a hypertree decomposition
type Node struct {
	ID             int
	Father         *Node
	Children       []*Node
	Bag            []string
	PossibleValues [][]int

	Lock *sync.Mutex
}

// AddChild to this node
func (node *Node) AddChild(child *Node) {
	node.Children = append(node.Children, child)
	child.Father = node
}
