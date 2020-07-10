package computation

import (
	. "../../CSP_Project/hyperTree"
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

func AttachPossibleSolutions(nodes []*Node) {
	wg := &sync.WaitGroup{}
	wg.Add(len(nodes))
	for _, node := range nodes {
		go attachSingleNode(node, wg)
	}
	wg.Wait()
}

func attachSingleNode(node *Node, wg *sync.WaitGroup) {
	defer wg.Done()
	file, err := os.Open("subCSP/" + strconv.Itoa(node.Id) + "sol.txt")
	defer file.Close()
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	var line string
	//reg := regexp.MustCompile("<instantiation id='sol\\d+' type='solution'> {2}<list>.*</list> {2}<values> (.*) </values> {2}</instantiation>.*")
	reg := regexp.MustCompile("v\\s+<instantiation>\\s+<list>.*</list>\\s+<values>(.*)</values>.*")
	for scanner.Scan() {
		line = scanner.Text()
		res := reg.FindStringSubmatch(line)
		if len(res) > 1 {
			value := make([]int, 0)
			for _, v := range strings.Split(res[1], " ") {
				i, _ := strconv.Atoi(v)
				value = append(value, i)
			}
			node.AddPossibleValue(value)
		}
	}
}
