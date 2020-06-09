package computation

import (
	. "../../CSP_Project/hyperTree"
)

func SequentialYannakaki(root *Node) {
	if len(root.Sons) != 0 {
		bottomUp(root)
		topDown(root)
	}
}

func bottomUp(actual *Node) {
	for _, son := range actual.Sons {
		bottomUp(son)
	}
	if actual.Father != nil {
		doSemiJoin(actual, actual.Father)
	}
}

func topDown(actual *Node) {
	for _, son := range actual.Sons {
		doSemiJoin(actual, son)
		topDown(son)
	}
}

// the left node performs the semi join on the right node and update the right's table
func doSemiJoin(left *Node, right *Node) {
	joiningIndex := make([][]int, 0) //TODO: migliorare il calcolo degli indici
	for iLeft, varLeft := range left.Variables {
		for iRight, varRight := range right.Variables {
			if varLeft == varRight {
				joiningIndex = append(joiningIndex, []int{iLeft, iRight})
				break
			}
		}
	}
	trashRow := make([]bool, len(right.PossibleValues)) //false at beginning
	for index, valuesRight := range right.PossibleValues {
		for _, valuesLeft := range left.PossibleValues {
			tupleMatch := true
			for _, rowIndex := range joiningIndex {
				if valuesLeft[rowIndex[0]] != valuesRight[rowIndex[1]] {
					tupleMatch = false
					break
				}
			}
			if tupleMatch {
				trashRow[index] = true
				break
			}
		}
	}
	for i := len(trashRow) - 1; i >= 0; i-- {
		if !trashRow[i] {
			right.PossibleValues = delByIndex(i, right.PossibleValues)
		}
	}
}

func delByIndex(index int, slice [][]int) [][]int {
	slice = append(slice[:index], slice[index+1:]...)
	return slice
}
