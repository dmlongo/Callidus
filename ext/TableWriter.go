package ext

import (
	"os"
	"strconv"
	"strings"

	"github.com/dmlongo/callidus/decomp"
)

// CreateSolutionTable creates a file containing the tuples of a given node
func CreateSolutionTable(tableFile string, node *decomp.Node) {
	table, err := os.OpenFile(tableFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		panic(err)
	}
	var sb strings.Builder
	for _, tup := range node.Tuples {
		for i, v := range tup {
			sb.WriteString(strconv.Itoa(v))
			if i < len(tup)-1 {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
		_, err := table.WriteString(sb.String())
		if err != nil {
			panic(err)
		}
		sb.Reset()
	}
}
