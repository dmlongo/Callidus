package decomp

import (
	"bufio"
	"io"
	"strings"
)

// Edge represents a hyperedge in a hypergraph
type Edge struct {
	name     string
	vertices []string
}

// Hypergraph is a collection of hyperedges
type Hypergraph map[string]Edge

// AddEdge to a hypergraph
func (hg Hypergraph) AddEdge(name string, vertices []string) {
	hg[name] = Edge{name: name, vertices: vertices}
}

// BuildHypergraph from a file
func BuildHypergraph(r *bufio.Reader) Hypergraph {
	hg := make(Hypergraph)
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF && len(line) == 0 {
			break
		}
		res := strings.Split(line, "(")
		name := res[0]
		vrts := res[1][:len(res[1])-3]
		vertices := strings.Split(vrts, ",")
		hg.AddEdge(name, vertices)
	}
	return hg
}
