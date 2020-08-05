package hyperTree

import "sync"

type Node struct {
	Id             int
	Variables      []string
	Father         *Node
	Sons           []*Node
	Lock           *sync.Mutex
	PossibleValues [][]int
}

func (node *Node) AddSon(node2 *Node) {
	node.Sons = append(node.Sons, node2)
	node2.Father = node
}

/*func (node *Node) SamePossibleValues(node2 *Node) bool {
	if node.Id != node2.Id {
		return false
	}
	if len(node.PossibleValues) != len(node2.PossibleValues) {
		return false
	}
	for i := range node.PossibleValues {
		if len(node.PossibleValues[i]) != len(node2.PossibleValues[i]) {
			return false
		}
		for j := range node.PossibleValues[i] {
			if node.PossibleValues[i][j] != node2.PossibleValues[i][j] {
				return false
			}
		}
	}
	return true
}*/
