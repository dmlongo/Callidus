package db

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/dmlongo/callidus/csp"
)

func RelToString(r Relation) string {
	var sb strings.Builder
	for i, attr := range r.Attributes() {
		sb.WriteString(attr)
		if i < len(r.Attributes()) {
			sb.WriteByte('\t')
		}
	}
	sb.WriteByte('\n')
	for _, tup := range r.Tuples() {
		for i, v := range tup {
			sb.WriteString(strconv.Itoa(v))
			if i < len(tup) {
				sb.WriteByte('\t')
			}
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

func RelToSolutions(r Relation) []csp.Solution {
	var res []csp.Solution
	for _, tup := range r.Tuples() {
		sol := make(csp.Solution)
		for i, v := range r.Attributes() {
			sol[v] = tup[i]
		}
		res = append(res, sol)
	}
	return res
}

// RelToFile creates a file containing the given relation
func RelToFile(filename string, r Relation) {
	table, err := os.Create(filename)
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
	for _, tup := range r.Tuples() {
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
