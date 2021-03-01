package decomp

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
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

func Yannakakis(root *Node, parallel bool, inMemory bool, folder string) *Node {
	joiningIndex := &MyMap{}
	joiningIndex.hash = make(map[IdJoiningIndex][][]int)
	joiningIndex.lock = &sync.RWMutex{}
	if len(root.Children) != 0 {
		if parallel {
			parallelBottomUp(root, joiningIndex, inMemory, folder)
			parallelTopDown(root, joiningIndex, inMemory, folder)
		} else {
			sequentialBottomUp(root, joiningIndex, inMemory, folder)
			sequentialTopDown(root, joiningIndex, inMemory, folder)
		}
	}
	return root
}

func sequentialBottomUp(actual *Node, joiningIndex *MyMap, inMemory bool, folder string) {
	for _, son := range actual.Children {
		sequentialBottomUp(son, joiningIndex, inMemory, folder)
	}
	if actual.Father != nil {
		doSemiJoin(actual, actual.Father, joiningIndex, inMemory, folder)
	}
}

func sequentialTopDown(actual *Node, joiningIndex *MyMap, inMemory bool, folder string) {
	for _, son := range actual.Children {
		doSemiJoin(actual, son, joiningIndex, inMemory, folder)
		sequentialTopDown(son, joiningIndex, inMemory, folder)
	}
}

func parallelBottomUp(actual *Node, joiningIndex *MyMap, inMemory bool, folder string) {
	var wg *sync.WaitGroup = &sync.WaitGroup{}
	for _, son := range actual.Children {
		if len(son.Children) != 0 && len(actual.Children) > 1 {
			wg.Add(1)
			go func(s *Node) {
				parallelBottomUp(s, joiningIndex, inMemory, folder)
				wg.Done()
			}(son)
		} else {
			parallelBottomUp(son, joiningIndex, inMemory, folder)
		}
	}
	wg.Wait()
	if actual.Father != nil {
		actual.Father.Lock.Lock()
		doSemiJoin(actual, actual.Father, joiningIndex, inMemory, folder)
		actual.Father.Lock.Unlock()
	}
}

func parallelTopDown(actual *Node, joiningIndex *MyMap, inMemory bool, folder string) {
	var wg *sync.WaitGroup = &sync.WaitGroup{}
	wg.Add(len(actual.Children))
	for _, son := range actual.Children {
		doSemiJoin(actual, son, joiningIndex, inMemory, folder)
		go func(s *Node) {
			parallelTopDown(s, joiningIndex, inMemory, folder)
			wg.Done()
		}(son)

	}
	wg.Wait()
}

// the left node performs the semi join on the right node and update the right's table
//TODO: potremmo cercare una correlazione nell'ordine in cui i semi-joins vengono effetuati
func doSemiJoin(left *Node, right *Node, joiningIndex *MyMap, inMemory bool, folder string) {
	indexJoin := searchJoiningIndex(left, right, joiningIndex)

	if inMemory {
		semiJoinInMemory(left, right, indexJoin)
	} else {
		semiJoinOnFile(left, right, indexJoin, folder)
	}

}

func semiJoinOnFile(left *Node, right *Node, indexJoin [][]int, folder string) {

	fileRight, rRight := OpenNodeFile(right.ID, folder)
	fileLeft, rLeft := OpenNodeFile(left.ID, folder)
	possibleValuesLeft := make([][]int, 0)
	for rLeft.Scan() {
		valuesLeft := GetValues(rLeft.Text(), len(left.Bag))
		if valuesLeft == nil {
			break
		}
		if valuesLeft[0] == -1 {
			return
		}
		possibleValuesLeft = append(possibleValuesLeft, valuesLeft)
	}
	fileLeft.Close()

	possibleValues := make([][]int, 0)

	for rRight.Scan() {
		valuesRight := GetValues(rRight.Text(), len(right.Bag))
		if valuesRight == nil {
			break
		}
		if valuesRight[0] == -1 {
			return
		}

		for _, valuesLeft := range possibleValuesLeft {
			tupleMatch := true
			for _, rowIndex := range indexJoin {
				if valuesLeft[rowIndex[0]] != valuesRight[rowIndex[1]] {
					tupleMatch = false
					break
				}
			}
			if tupleMatch {
				possibleValues = append(possibleValues, valuesRight)
				break
			}
		}
	}

	fileRight.Close()

	fileRight, err := os.OpenFile("tables-"+folder+strconv.Itoa(right.ID)+".table", os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		panic(err)
	}

	if len(possibleValues) == 0 {
		fileRight.WriteString("-1")
	} else {
		for _, row := range possibleValues {
			for i, val := range row {
				if i == len(row)-1 {
					fileRight.WriteString(strconv.Itoa(val))
				} else {
					fileRight.WriteString(strconv.Itoa(val) + " ")
				}

			}
			fileRight.WriteString("\n")
		}
	}

	fileRight.Close()
}

func semiJoinInMemory(left *Node, right *Node, indexJoin [][]int) {
	trashRow := make([]bool, len(right.PossibleValues))
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

func searchJoiningIndex(left *Node, right *Node, joiningIndex *MyMap) [][]int {
	indexJoin := make([][]int, 0)
	joiningIndex.lock.RLock()
	val, ok := joiningIndex.hash[IdJoiningIndex{x: left.ID, y: right.ID}]
	joiningIndex.lock.RUnlock()
	if ok {
		indexJoin = val
	} else {
		invertedIndex := make([][]int, 0)
		for iLeft, varLeft := range left.Bag {
			for iRight, varRight := range right.Bag {
				if varLeft == varRight {
					invertedIndex = append(invertedIndex, []int{iRight, iLeft})
					indexJoin = append(indexJoin, []int{iLeft, iRight})
					break
				}
			}
		}
		joiningIndex.lock.Lock()
		joiningIndex.hash[IdJoiningIndex{x: right.ID, y: left.ID}] = invertedIndex
		joiningIndex.lock.Unlock()
	}
	return indexJoin
}

func OpenNodeFile(id int, folder string) (*os.File, *bufio.Scanner) {
	fi, err := os.Open("tables-" + folder + strconv.Itoa(id) + ".table")
	if err != nil {
		panic(err)
	}
	err = fi.Chmod(0777)
	if err != nil {
		panic(err)
	}
	return fi, bufio.NewScanner(fi)
}

func GetValues(line string, numVariables int) []int {
	reg := regexp.MustCompile("(.*)")
	valuesString := reg.FindStringSubmatch(line)[1]
	if valuesString == "" {
		return nil
	}
	if valuesString == "-1" {
		return []int{-1}
	}
	values := make([]int, numVariables)
	for i, value := range strings.Split(valuesString, " ") {
		v, err := strconv.Atoi(value)
		if err != nil {
			panic(err)
		}
		values[i] = v

	}
	return values
}
