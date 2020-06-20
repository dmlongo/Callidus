package computation

import (
	. "../../CSP_Project/hyperTree"
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

func delByIndex(index int, slice [][]int) [][]int {
	if index+1 >= len(slice) {
		slice = slice[:index]
	} else {
		slice = append(slice[:index], slice[index+1:]...)
	}
	return slice
}

//TODO: forse sequential e parallel possono essere accorpati in caso dovremo utilizzarli parecchio
func SequentialYannakaki(root *Node) *Node {
	joiningIndex := &MyMap{}
	joiningIndex.hash = make(map[IdJoiningIndex][][]int)
	joiningIndex.lock = &sync.RWMutex{}
	if len(root.Sons) != 0 {
		sequentialBottomUp(root, joiningIndex)
		sequentialTopDown(root, joiningIndex)
	}
	return root
}

func sequentialBottomUp(actual *Node, joiningIndex *MyMap) {
	for _, son := range actual.Sons {
		sequentialBottomUp(son, joiningIndex)
	}
	if actual.Father != nil {
		doSemiJoin(actual, actual.Father, joiningIndex)
	}
}

func sequentialTopDown(actual *Node, joiningIndex *MyMap) {
	for _, son := range actual.Sons {
		doSemiJoin(actual, son, joiningIndex)
		sequentialTopDown(son, joiningIndex)
	}
}

func ParallelYannakaki(root *Node) *Node {
	joiningIndex := &MyMap{}
	joiningIndex.hash = make(map[IdJoiningIndex][][]int)
	joiningIndex.lock = &sync.RWMutex{}
	if len(root.Sons) != 0 {
		parallelBottomUp(root, joiningIndex)
		parallelTopDown(root, joiningIndex)
	}
	return root
}

func parallelBottomUp(actual *Node, joiningIndex *MyMap) {
	var wg sync.WaitGroup
	for _, son := range actual.Sons {
		s := son
		if len(son.Sons) != 0 {
			wg.Add(1)
			go func() {
				parallelBottomUp(s, joiningIndex)
				wg.Done()
			}()
		} else {
			parallelBottomUp(s, joiningIndex)
		}
	}
	wg.Wait()
	if actual.Father != nil {
		doSemiJoin(actual, actual.Father, joiningIndex)
	}
}

func parallelTopDown(actual *Node, joiningIndex *MyMap) {
	var wg sync.WaitGroup
	wg.Add(len(actual.Sons))
	for _, son := range actual.Sons {
		s := son
		doSemiJoin(actual, s, joiningIndex)
		go func() {
			parallelTopDown(s, joiningIndex)
			wg.Done()
		}()
	}
	wg.Wait()
}

// the left node performs the semi join on the right node and update the right's table
//TODO: potremmo cercare una correlazione nell'ordine in cui i semi-joins vengono effetuati
func doSemiJoin(left *Node, right *Node, joiningIndex *MyMap) {
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
}
