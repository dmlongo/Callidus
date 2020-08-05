package computation

import (
	. "../../Callidus/hyperTree"
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

func Yannakaki(root *Node, version bool, folderName string) *Node {
	joiningIndex := &MyMap{}
	joiningIndex.hash = make(map[IdJoiningIndex][][]int)
	joiningIndex.lock = &sync.RWMutex{}
	if len(root.Sons) != 0 {
		if version {
			parallelBottomUp(root, joiningIndex, folderName)
			parallelTopDown(root, joiningIndex, folderName)
		} else {
			sequentialBottomUp(root, joiningIndex, folderName)
			sequentialTopDown(root, joiningIndex, folderName)
		}
	}
	return root
}

func sequentialBottomUp(actual *Node, joiningIndex *MyMap, folderName string) {
	for _, son := range actual.Sons {
		sequentialBottomUp(son, joiningIndex, folderName)
	}
	if actual.Father != nil {
		doSemiJoin(actual, actual.Father, joiningIndex, folderName)
	}
}

func sequentialTopDown(actual *Node, joiningIndex *MyMap, folderName string) {
	for _, son := range actual.Sons {
		doSemiJoin(actual, son, joiningIndex, folderName)
		sequentialTopDown(son, joiningIndex, folderName)
	}
}

func parallelBottomUp(actual *Node, joiningIndex *MyMap, folderName string) {
	var wg *sync.WaitGroup = &sync.WaitGroup{}
	for _, son := range actual.Sons {
		if len(son.Sons) != 0 && len(actual.Sons) > 1 {
			wg.Add(1)
			go func(s *Node) {
				parallelBottomUp(s, joiningIndex, folderName)
				wg.Done()
			}(son)
		} else {
			parallelBottomUp(son, joiningIndex, folderName)
		}
	}
	wg.Wait()
	if actual.Father != nil {
		actual.Father.Lock.Lock()
		doSemiJoin(actual, actual.Father, joiningIndex, folderName)
		actual.Father.Lock.Unlock()
	}
}
func parallelTopDown(actual *Node, joiningIndex *MyMap, folderName string) {
	var wg *sync.WaitGroup = &sync.WaitGroup{}
	wg.Add(len(actual.Sons))
	for _, son := range actual.Sons {
		doSemiJoin(actual, son, joiningIndex, folderName)
		go func(s *Node) {
			parallelTopDown(s, joiningIndex, folderName)
			wg.Done()
		}(son)

	}
	wg.Wait()
}

// the left node performs the semi join on the right node and update the right's table
//TODO: potremmo cercare una correlazione nell'ordine in cui i semi-joins vengono effetuati
func doSemiJoin(left *Node, right *Node, joiningIndex *MyMap, folderName string) {
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

	fileRight, rRight := OpenNodeFile(right.Id, folderName)
	fileLeft, rLeft := OpenNodeFile(left.Id, folderName)

	defer fileRight.Close()
	defer fileLeft.Close()

	possibleValues := make([][]int, 0)
	for rRight.Scan() {
		valuesRight := GetValues(rRight, len(right.Variables))
		for rLeft.Scan() {
			valuesLeft := GetValues(rLeft, len(left.Variables))
			tupleMatch := true
			for _, rowIndex := range indexJoin {
				if valuesLeft[rowIndex[0]] != valuesRight[rowIndex[1]] {
					tupleMatch = false
					break
				}
			}
			if tupleMatch {
				possibleValues = append(possibleValues, valuesLeft)
				break
			}
		}
	}

	writer := bufio.NewWriter(fileLeft)

	for _, row := range possibleValues {
		for _, val := range row {
			writer.WriteString(strconv.Itoa(val))
		}
		writer.WriteString("\n")
	}

	//for index, valuesRight := range right.PossibleValues {
	//	for _, valuesLeft := range left.PossibleValues {
	//		tupleMatch := true
	//		for _, rowIndex := range indexJoin {
	//			if valuesLeft[rowIndex[0]] != valuesRight[rowIndex[1]] {
	//				tupleMatch = false
	//				break
	//			}
	//		}
	//		if tupleMatch {
	//			trashRow[index] = true
	//			break
	//		}
	//	}
	//}
	//for i := len(trashRow) - 1; i >= 0; i-- {
	//	if !trashRow[i] {
	//		right.PossibleValues = delByIndex(i, right.PossibleValues)
	//	}
	//}
}

func OpenNodeFile(id int, folderName string) (*os.File, *bufio.Scanner) {
	fi, err := os.Open("tables-" + folderName + "/" + strconv.Itoa(id) + ".table")
	if err != nil {
		panic(err)
	}
	err = fi.Chmod(0777)
	if err != nil {
		panic(err)
	}
	return fi, bufio.NewScanner(fi)
}

func GetValues(scanner *bufio.Scanner, numVariables int) []int {
	reg := regexp.MustCompile("(\\d*)")
	line := scanner.Text()
	values := make([]int, numVariables)
	for i, value := range strings.Split(reg.FindStringSubmatch(line)[1], " ") {
		v, err := strconv.Atoi(value)
		if err != nil {
			panic(err)
		}
		values[i] = v
	}
	return values
}
