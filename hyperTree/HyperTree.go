package hyperTree

import "sync"

type Node struct {
	Id             int
	Variables      []string
	Father         *Node
	Sons           []*Node
	Lock           *sync.Mutex
	PossibleValues map[string][]string
}

func (node *Node) AddSon(node2 *Node) {
	node.Sons = append(node.Sons, node2)
}
