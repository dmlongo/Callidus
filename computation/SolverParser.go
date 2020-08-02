package computation

import (
	. "../../Callidus/hyperTree"
	"bufio"
	"os"
	"strconv"
	"strings"
)

func AttachSingleNode(folderName string, node *Node, debugOption bool) {
	file, err := os.Open(folderName + strconv.Itoa(node.Id) + "sol.txt")
	if err != nil {
		panic(err)
	}
	defer func(debugOption bool) {
		if !debugOption {
			err := os.Remove(file.Name())
			if err != nil {
				panic(err)
			}
		}
	}(debugOption)
	scanner := bufio.NewScanner(file)
	var line string
	node.PossibleValues = make([][]int, 0)
	for scanner.Scan() {
		line = scanner.Text()
		values := strings.Split(line, " ")
		var temp []int
		for _, v := range values {
			value, err := strconv.Atoi(v)
			if err != nil {
				panic(err)
			}
			temp = append(temp, value)
		}
		node.PossibleValues = append(node.PossibleValues, temp)
	}
	err = file.Close()
	if err != nil {
		panic(err)
	}
}
