package ext

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/dmlongo/callidus/decomp"
)

// CreateSolutionTable creates a file containing the tuples of a given node
func CreateSolutionTable(tableFile string, node *decomp.Node) {
	table, err := os.Create(tableFile)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := table.Close(); err != nil {
			panic(err)
		}
	}()

	w := bufio.NewWriter(table)
	var sb strings.Builder
	for _, tup := range node.Tuples {
		for i, v := range tup {
			sb.WriteString(strconv.Itoa(v))
			if i < len(tup)-1 {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
		_, err := w.WriteString(sb.String())
		if err != nil {
			panic(err)
		}
		sb.Reset()
	}
}
