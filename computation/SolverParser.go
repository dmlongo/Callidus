package computation

import (
	. "../../Callidus/hyperTree"
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func AttachSingleNode(folderName string, node *Node, debugOption bool) {
	defer func(debugOption bool) {
		if !debugOption {
			err := os.RemoveAll(folderName + strconv.Itoa(node.Id) + "sol.txt")
			if err != nil {
				panic(err)
			}
		}
	}(debugOption)
	file, err := os.Open(folderName + strconv.Itoa(node.Id) + "sol.txt")
	if err != nil {
		panic(err)
	}
	defer func(debugOption bool) {
		if !debugOption {
			err := os.RemoveAll(file.Name())
			if err != nil {
				panic(err)
			}
		}
	}(debugOption)
	scanner := bufio.NewScanner(file)
	var line string
	node.PossibleValues = make(map[string][]string)
	for scanner.Scan() {
		line = scanner.Text()
		reg := regexp.MustCompile("(.*) -> (.*)")
		variable := reg.FindStringSubmatch(line)[1]
		values := strings.Split(reg.FindStringSubmatch(line)[2], " ")
		node.PossibleValues[variable] = values
	}
	err = file.Close()
	if err != nil {
		panic(err)
	}
}
