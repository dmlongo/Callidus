package decomp

import (
	"bufio"
	"regexp"
	"strings"

	files "github.com/dmlongo/callidus/ext/files"
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

var edgeRegex = regexp.MustCompile(`(\w+)\((\w+(,\w+)*)\)[,.]`)

// BuildHypergraph from a file
func BuildHypergraph(r *bufio.Reader) Hypergraph {
	hg := make(Hypergraph)
	for {
		line, eof := files.ReadLine(r)
		if eof {
			break
		}
		line = strings.ReplaceAll(line, " ", "")
		matches := edgeRegex.FindStringSubmatch(line)
		if len(matches) < 3 {
			panic("Bad edge= " + line)
		}
		name := matches[1]
		vertices := strings.Split(matches[2], ",")
		hg.AddEdge(name, vertices)
	}
	return hg
}
