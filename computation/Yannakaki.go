package computation

import (
	. "../../Callidus/hyperTree"
	"sync"
)

func delByIndex(index int, slice []string) []string {
	if index+1 >= len(slice) {
		slice = slice[:index]
	} else {
		slice = append(slice[:index], slice[index+1:]...)
	}
	return slice
}

func Yannakaki(root *Node, version bool) *Node {
	if len(root.Sons) != 0 {
		if version {
			parallelBottomUp(root)
			parallelTopDown(root)
		} else {
			sequentialBottomUp(root)
			sequentialTopDown(root)
		}
	}
	return root
}

func sequentialBottomUp(actual *Node) {
	for _, son := range actual.Sons {
		sequentialBottomUp(son)
	}
	if actual.Father != nil {
		doSemiJoin(actual, actual.Father)
	}
}

func sequentialTopDown(actual *Node) {
	for _, son := range actual.Sons {
		doSemiJoin(actual, son)
		sequentialTopDown(son)
	}
}

func parallelBottomUp(actual *Node) {
	var wg = &sync.WaitGroup{}
	for _, son := range actual.Sons {
		if len(son.Sons) != 0 && len(actual.Sons) > 1 {
			wg.Add(1)
			go func(s *Node) {
				parallelBottomUp(s)
				wg.Done()
			}(son)
		} else {
			parallelBottomUp(son)
		}
	}
	wg.Wait()
	if actual.Father != nil {
		actual.Father.Lock.Lock()
		doSemiJoin(actual, actual.Father)
		actual.Father.Lock.Unlock()
	}
}

func parallelTopDown(actual *Node) {
	for _, son := range actual.Sons {
		doSemiJoin(actual, son)
		go parallelTopDown(son)
	}
}

func doSemiJoin(left *Node, right *Node) {
	//indexToKeep := make([]bool, len(right.PossibleValues[right.Variables[0]])) //false at beginning
	indexToDiscard := make(map[int]struct{})
	for _, variable := range left.Variables {
		if _, join := right.PossibleValues[variable]; join {
			leftValues := left.PossibleValues[variable]
			rightValues := right.PossibleValues[variable]
			for i, value := range rightValues {
				if !isElementInSlice(value, leftValues) {
					indexToDiscard[i] = struct{}{}
				}
			}
		}
	}
	if len(indexToDiscard) > 0 {
		newMap := make(map[string][]string)
		for key, values := range right.PossibleValues {
			newValues := make([]string, len(values)-len(indexToDiscard))
			newIndex := 0
			for index := range values {
				if _, exist := indexToDiscard[index]; !exist {
					newValues[newIndex] = values[index]
					newIndex++
				}
			}
			newMap[key] = newValues
		}
		right.PossibleValues = newMap
	}
	/*for _, possibleValues := range right.PossibleValues {
		var newPossibleValues []string
		for i, value := range indexToKeep {
			if value {
				newPossibleValues = append(newPossibleValues, possibleValues[i])
			} else {
				fmt.Println("asdfdghj")
			}
		}
		if len(possibleValues) != len(newPossibleValues) {
			fmt.Println("fesdfghj")
		}
		possibleValues = newPossibleValues
	}*/
}

func isElementInSlice(value string, slice []string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// the left node performs the semi join on the right node and update the right's table
//TODO: potremmo cercare una correlazione nell'ordine in cui i semi-joins vengono effetuati
/*func doSemiJoin(left *Node, right *Node, joiningIndex *MyMap) {
	indexJoin := make([][]int, 0)
	joiningIndex.lock.RLock()
	val, ok := joiningIndex.hash[IdJoiningIndex{x: left.Id, y: right.Id}]
	joiningIndex.lock.RUnlock()
	if ok {
		indexJoin = val
	} else {
		invertedIndex := make([][]int, 0)
		for iLeft, varLeft := range left.Variables {
			for iRight, varRight := range right.Variables {
				if varLeft == varRight {
					invertedIndex = append(invertedIndex, []int{iRight, iLeft})
					indexJoin = append(indexJoin, []int{iLeft, iRight})
					break
				}
			}
		}
		joiningIndex.lock.Lock()
		joiningIndex.hash[IdJoiningIndex{x: right.Id, y: left.Id}] = invertedIndex
		joiningIndex.lock.Unlock()
	}
	trashRow := make([]bool, len(right.PossibleValues)) //false at beginning
	for index, valuesRight := range right.PossibleValues {
		for _, valuesLeft := range left.PossibleValues {
			tupleMatch := true
			for _, rowIndex := range indexJoin {
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
}*/
