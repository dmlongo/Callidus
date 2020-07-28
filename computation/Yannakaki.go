package computation

import (
	. "../../Callidus/hyperTree"
	"sync"
)

type IdJoiningIndex struct {
	x int
	y int
}

type MyMap struct {
	hash map[IdJoiningIndex][][]int
	lock *sync.RWMutex
}

func delByIndex(index int, slice []string) []string {
	if index+1 >= len(slice) {
		slice = slice[:index]
	} else {
		slice = append(slice[:index], slice[index+1:]...)
	}
	return slice
}

func Yannakaki(root *Node, version bool) *Node {
	joiningIndex := &MyMap{}
	joiningIndex.hash = make(map[IdJoiningIndex][][]int)
	joiningIndex.lock = &sync.RWMutex{}
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
	trashIndex := make([]bool, len(right.PossibleValues[right.Variables[0]])) //false at beginning
	for leftVariable, leftValues := range left.PossibleValues {
		if rightValues, check := right.PossibleValues[leftVariable]; check {
			for index, values := range rightValues {
				if !isElementInSlice(values, leftValues) {
					trashIndex[index] = true
				}
			}
		}
	}
	for _, possibleValues := range right.PossibleValues {
		for i := len(trashIndex) - 1; i >= 0; i-- {
			if !trashIndex[i] {
				possibleValues = delByIndex(i, possibleValues)
			}
		}
	}
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
