package computation

import (
	. "../../Callidus/hyperTree"
	"bufio"
	"io"
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
	//fmt.Println(left.Id, right.Id)
	//fmt.Println("Joining index", indexJoin)
	//fileRight, rRight := OpenNodeFile(right.Id, folderName)
	//fileLeft, rLeft := OpenNodeFile(left.Id, folderName)
	//for rLeft.Scan() {
	//	valuesLeft := GetValues(rLeft, len(left.Variables))
	//	fmt.Println(valuesLeft)
	//}
	//fmt.Println("\n")
	//for rRight.Scan() {
	//	valuesRight := GetValues(rRight, len(right.Variables))
	//	fmt.Println(valuesRight)
	//}
	//
	//fileLeft.Close()
	//fileRight.Close()

	fileRight, rRight := OpenNodeFile(right.Id, folderName)
	fileLeft, rLeft := OpenNodeFile(left.Id, folderName)

	possibleValues := make([][]int, 0)

	for rRight.Scan() {
		valuesRight := GetValues(rRight.Text(), len(right.Variables))
		//fmt.Println(valuesRight)
		if valuesRight == nil {
			break
		}
		if valuesRight[0] == -1 {
			return
		}

		fileLeft.Seek(0, io.SeekStart)
		rLeft = bufio.NewScanner(fileLeft)
		i := 0
		for rLeft.Scan() {
			//fmt.Println(fileLeft.Name(), i)
			valuesLeft := GetValues(rLeft.Text(), len(left.Variables))
			if valuesLeft == nil {
				break
			}
			if valuesLeft[0] == -1 {
				return
			}
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
			i++
		}
	}
	//fmt.Println(possibleValues)
	//fmt.Println("\n\n")

	fileRight.Close()
	fileLeft.Close()

	fileRight, err := os.OpenFile("tables-"+folderName+strconv.Itoa(right.Id)+".table", os.O_TRUNC|os.O_WRONLY, 0777)
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
	fi, err := os.Open("tables-" + folderName + strconv.Itoa(id) + ".table")
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
	//fmt.Println(line)
	reg := regexp.MustCompile("(.*)")
	valuesString := reg.FindStringSubmatch(line)[1]
	if valuesString == "" {
		return nil
	}
	if valuesString == "-1" {
		return []int{-1}
	}
	values := make([]int, numVariables)
	//if line == "8 9 5 1 3 6 2 4 8 7"{
	//	fmt.Println(numVariables)
	//	fmt.Println(valuesString)
	//	fmt.Println(strings.Split(valuesString, " "))
	//}

	//fmt.Println(numVariables, len(strings.Split(valuesString, " ")), line)

	for i, value := range strings.Split(valuesString, " ") {
		//if line == "8 9 5 1 3 4 6 2 8 7" {
		//	fmt.Print(value, "index",i)
		//}
		v, err := strconv.Atoi(value)
		if err != nil {
			panic(err)
		}
		values[i] = v
		//if line == "8 9 5 1 3 4 6 2 8 7" {
		//	fmt.Println(values[i])
		//}

	}
	//fmt.Println("\n")
	return values
}
