package decomp

import (
	"fmt"
	"strings"
	"testing"
)

func TestBfs(t *testing.T) {
	n1 := &Node{ID: 1}
	n2 := &Node{ID: 2}
	n3 := &Node{ID: 3}
	n4 := &Node{ID: 4}
	n5 := &Node{ID: 5}
	n6 := &Node{ID: 6}
	n7 := &Node{ID: 7}
	n1.AddChild(n2)
	n1.AddChild(n3)
	n1.AddChild(n6)
	n3.AddChild(n4)
	n3.AddChild(n5)
	n6.AddChild(n7)

	expected := []int{1, 2, 3, 6, 4, 5, 7}

	var result []int
	for _, n := range Bfs(n1) {
		result = append(result, n.ID)
	}

	if len(result) != len(expected) {
		t.Error(fmt.Sprintf("%s\nlen(result) = %v, len(expected) = %v", print(result, expected), len(result), len(expected)))
	}
	for i := range result {
		if result[i] != expected[i] {
			t.Error(fmt.Sprintf("%s\nresult[%v] = %v, expected[%v] = %v", print(result, expected), i, result[i], i, expected[i]))
			t.Error(print(result, expected) + "result[i] != expected[i]")
		}
	}
}

func print(res []int, exp []int) string {
	var sb strings.Builder
	sb.WriteString("res= ")
	sb.WriteString(fmt.Sprintf("%v\t", res))
	sb.WriteString("exp= ")
	sb.WriteString(fmt.Sprintf("%v", exp))
	return sb.String()
}
