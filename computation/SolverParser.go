package computation

import (
	. "../../CSP_Project/hyperTree"
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func AttachPossibleSolutions(nodes []*Node) bool {
	exit := make(chan bool, 100)
	defer close(exit)
	for _, node := range nodes {
		go attachSingleNode(node, &exit)
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

func attachSingleNode(node *Node, exit *chan bool) {
	file, err := os.Open("subCSP/" + strconv.Itoa(node.Id) + "sol.txt")
	defer file.Close()
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	var line string
	//reg := regexp.MustCompile("<instantiation id='sol\\d+' type='solution'> {2}<list>.*</list> {2}<values> (.*) </values> {2}</instantiation>.*")
	regSol := regexp.MustCompile("v\\s+<instantiation>\\s+<list>.*</list>\\s+<values>(.*)</values>.*")
	regNumSol := regexp.MustCompile("c # Sols = (.*)")
	for scanner.Scan() {
		line = scanner.Text()
		res := regSol.FindStringSubmatch(line)
		if len(res) > 1 {
			value := make([]int, 0)
			for _, v := range strings.Split(res[1], " ") {
				i, _ := strconv.Atoi(v)
				value = append(value, i)
			}
			node.AddPossibleValue(value)
		} else {
			res := regNumSol.FindStringSubmatch(line)
			if len(res) > 1 {
				num := strings.Split(res[1], " ")
				if num[0] == "0" {
					*exit <- false
					return
				}
			}
		}
	}
	*exit <- true
}
