package computation

import (
	. "../../Callidus/hyperTree"
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func AttachPossibleSolutions(folderName string, nodes []*Node) bool {
	exit := make(chan bool, 100)
	defer close(exit)
	for _, node := range nodes {
		attachSingleNode(folderName, node, &exit)
	}
	cont := 0
	for {
		select {
		case res := <-exit:
			if res {
				cont++
				if cont == len(nodes) {
					return true
				}
			} else {
				return false
			}
		}
	}
}

func attachSingleNode(folderName string, node *Node, exit *chan bool) {
	file, err := os.Open(folderName + strconv.Itoa(node.Id) + "sol.txt")
	if err != nil {
		panic(err)
	}
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
	fmt.Print(strconv.Itoa(node.Id) + " ->")
	fmt.Println(node.PossibleValues)
	err = file.Close()
	if err != nil {
		panic(err)
	}
	*exit <- true
}
