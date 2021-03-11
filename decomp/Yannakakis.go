package decomp

import (
	"sync"
)

// IDJoiningIndex is.. I still don't know
type IDJoiningIndex struct {
	x int
	y int
}

// MyMap is a map with a lock
type MyMap struct {
	hash map[IDJoiningIndex][][]int
	lock *sync.RWMutex
}

// YannakakisSeq performs the sequential Yannakakis' algorithm
func YannakakisSeq(root *Node) *Node {
	joiningIndex := &MyMap{}
	joiningIndex.hash = make(map[IDJoiningIndex][][]int)
	joiningIndex.lock = &sync.RWMutex{}

	if len(root.Children) != 0 {
		bottomUpSeq(root, joiningIndex)
		topDownSeq(root, joiningIndex)
	}
	return root
}

func bottomUpSeq(parent *Node, joiningIndex *MyMap) {
	for _, child := range parent.Children {
		bottomUpSeq(child, joiningIndex)
	}
	if parent.Parent != nil {
		semiJoin(parent.Parent, parent, joiningIndex)
	}
}

func topDownSeq(parent *Node, joiningIndex *MyMap) {
	for _, child := range parent.Children {
		semiJoin(child, parent, joiningIndex)
		topDownSeq(child, joiningIndex)
	}
}

func semiJoin(left *Node, right *Node, joiningIndex *MyMap) {
	indexJoin := searchJoiningIndex(left, right, joiningIndex)

	var tupToDel []int
	for i, leftTup := range left.Tuples {
		delete := true
		for _, rightTup := range right.Tuples {
			if match(leftTup, rightTup, indexJoin) {
				delete = false
				break
			}
		}
		if delete {
			tupToDel = append(tupToDel, i)
		}
	}

	update(left, tupToDel)
}

func match(left []int, right []int, joinIndex [][]int) bool {
	for _, z := range joinIndex {
		if left[z[0]] != right[z[1]] {
			return false
		}
	}
	return true
}

func update(n *Node, toDel []int) { // TODO shouldn't this method be synchronized?
	if len(toDel) > 0 {
		newSize := len(n.Tuples) - len(toDel)
		newTuples := make(Relation, 0, newSize)
		if newSize > 0 { // TODO what does newSize == 0 mean? unsat?
			i := 0
			for _, j := range toDel {
				newTuples = append(newTuples, n.Tuples[i:j]...)
				i = j + 1
			}
			newTuples = append(newTuples, n.Tuples[i:]...)
		}
		n.Tuples = newTuples
	}
}

// the left node performs the semi join on the right node and update the right's table
//TODO: potremmo cercare una correlazione nell'ordine in cui i semi-joins vengono effetuati
func semiJoinOld(left *Node, right *Node, joiningIndex *MyMap) {
	indexJoin := searchJoiningIndex(left, right, joiningIndex)

	trashRow := make([]bool, len(right.Tuples))
	for index, valuesRight := range right.Tuples {
		for _, valuesLeft := range left.Tuples {
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
			right.Tuples = delByIndex(i, right.Tuples)
		}
	}
}

func delByIndex(index int, slice [][]int) [][]int {
	if index+1 >= len(slice) {
		slice = slice[:index]
	} else {
		slice = append(slice[:index], slice[index+1:]...)
	}
	return slice
}

func searchJoiningIndex(left *Node, right *Node, joiningIndex *MyMap) [][]int {
	var joinIndices [][]int
	joiningIndex.lock.RLock()
	val, ok := joiningIndex.hash[IDJoiningIndex{x: left.ID, y: right.ID}]
	joiningIndex.lock.RUnlock()
	if ok {
		joinIndices = val
	} else {
		var invJoinIndices [][]int
		for iLeft, varLeft := range left.Bag {
			if iRight := right.Position(varLeft); iRight >= 0 {
				joinIndices = append(joinIndices, []int{iLeft, iRight})
				invJoinIndices = append(invJoinIndices, []int{iRight, iLeft})
			}
		}
		joiningIndex.lock.Lock()
		joiningIndex.hash[IDJoiningIndex{x: right.ID, y: left.ID}] = invJoinIndices
		joiningIndex.lock.Unlock()
	}
	return joinIndices
}

func searchJoiningIndexOld(left *Node, right *Node, joiningIndex *MyMap) [][]int {
	indexJoin := make([][]int, 0)
	joiningIndex.lock.RLock()
	val, ok := joiningIndex.hash[IDJoiningIndex{x: left.ID, y: right.ID}]
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
		joiningIndex.hash[IDJoiningIndex{x: right.ID, y: left.ID}] = invertedIndex
		joiningIndex.lock.Unlock()
	}
	return indexJoin
}

/*
func semiJoinInMemory(left *Node, right *Node, indexJoin [][]int) {
	trashRow := make([]bool, len(right.Tuples))
	for index, valuesRight := range right.Tuples {
		for _, valuesLeft := range left.Tuples {
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
			right.Tuples = delByIndex(i, right.Tuples)
		}
	}
}
*/

// YannakakisPar performs the parrallel Yannakakis' algorithm
func YannakakisPar(root *Node) *Node {
	joiningIndex := &MyMap{}
	joiningIndex.hash = make(map[IDJoiningIndex][][]int)
	joiningIndex.lock = &sync.RWMutex{}

	if len(root.Children) != 0 {
		parallelBottomUp(root, joiningIndex)
		parallelTopDown(root, joiningIndex)
	}
	return root
}

func parallelBottomUp(actual *Node, joiningIndex *MyMap) {
	var wg *sync.WaitGroup = &sync.WaitGroup{}
	for _, son := range actual.Children {
		if len(son.Children) != 0 && len(actual.Children) > 1 {
			wg.Add(1)
			go func(s *Node) {
				parallelBottomUp(s, joiningIndex)
				wg.Done()
			}(son)
		} else {
			parallelBottomUp(son, joiningIndex)
		}
	}
	wg.Wait()
	if actual.Parent != nil {
		actual.Parent.Lock.Lock()
		semiJoin(actual, actual.Parent, joiningIndex)
		actual.Parent.Lock.Unlock()
	}
}

func parallelTopDown(actual *Node, joiningIndex *MyMap) {
	var wg *sync.WaitGroup = &sync.WaitGroup{}
	wg.Add(len(actual.Children))
	for _, son := range actual.Children {
		semiJoin(actual, son, joiningIndex)
		go func(s *Node) {
			parallelTopDown(s, joiningIndex)
			wg.Done()
		}(son)

	}
	wg.Wait()
}

/*
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
*/

/*
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
*/

/*
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
*/
