package computation

import (
	. "../../Callidus/hyperTree"
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

func AttachPossibleSolutions(folderName string, nodes []*Node) bool {
	exit := make(chan bool, 100)
	defer close(exit)
	for _, node := range nodes {
		go attachSingleNode(folderName, node, &exit)
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
	return true
}

func attachSingleNode(folderName string, node *Node, exit *chan bool) {
	file, err := os.Open(folderName + strconv.Itoa(node.Id) + "sol.txt")
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	var line string
	for scanner.Scan() {
		line = scanner.Text()
		keyReg := regexp.MustCompile("(.*) ->.*")
		variable := keyReg.FindStringSubmatch(line)
		valueReg := regexp.MustCompile(".* -> (.*)") //\n?
		values := valueReg.FindStringSubmatch(line)
		fmt.Print(variable[0] + " -> ")
		fmt.Println(values[0])
	}
	err = file.Close()
	if err != nil {
		panic(err)
	}
	*exit <- true
}
