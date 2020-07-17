package computation

import (
	. "../../Callidus/hyperTree"
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func AttachPossibleSolutions(nodes []*Node, solutions *[]string, inMemory bool, solver string) bool {
	exit := make(chan bool, 100)
	defer close(exit)
	regSol, regNumSol := parseSolver(solver)
	for i, node := range nodes {
		if inMemory {
			sol := (*solutions)[i]
			go attachSingleNodeInMemory(node, &exit, &sol, regSol, regNumSol)
		} else {
			go attachSingleNode(node, &exit, regSol, regNumSol)
		}

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

func attachSingleNode(node *Node, exit *chan bool, regSol *regexp.Regexp, regNumSol *regexp.Regexp) {
	file, err := os.Open("subCSP/" + strconv.Itoa(node.Id) + "sol.txt")
	defer file.Close()
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	var line string
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

func attachSingleNodeInMemory(node *Node, exit *chan bool, solution *string, regSol *regexp.Regexp, regNumSol *regexp.Regexp) {
	output := strings.Split(*solution, "\n")
	for _, line := range output {
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

func parseSolver(solver string) (*regexp.Regexp, *regexp.Regexp) {
	switch solver {
	case "Nacre":
		regSol := regexp.MustCompile("v\\s+<instantiation>\\s+<list>.*</list>\\s+<values>(.*) </values>.*")
		regNumSol := regexp.MustCompile("c # Sols = (.*)")
		return regSol, regNumSol
	case "AbsCon":
		regSol := regexp.MustCompile("<instantiation id='sol\\d+' type='solution'> {2}<list>.*</list> {2}<values> (.*) </values> {2}</instantiation>.*")
		regNumSol := regexp.MustCompile("<nbSolutions>\\s+(.*)\\s+</nbSolutions>")
		return regSol, regNumSol
	}
	panic("solver not found")
	return nil, nil
}
