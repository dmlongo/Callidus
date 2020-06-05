package hyperTree

type Node struct {
	id        int
	joinNodes []string
	variables []string
	father    *Node
	sons      *[]Node
}
