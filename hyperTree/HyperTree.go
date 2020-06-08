package hyperTree

type Node struct {
	Id             int
	JoinNodes      []string
	Variables      []string
	Father         *Node
	Sons           []*Node
	PossibleValues [][]int
}

func (node *Node) AddSon(node2 *Node) {
	node.Sons = append(node.Sons, node2)
}

func (node *Node) AddFather(node2 *Node) {
	node.Father = node2
}

func (node *Node) AddPossibleValue(value []int) {
	node.PossibleValues = append(node.PossibleValues, value)
}
