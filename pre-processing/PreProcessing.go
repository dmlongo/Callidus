package pre_processing

import (
	. "../../CSP_Project/constraint"
	. "../../CSP_Project/hyperTree"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

func HypergraphTranslation(filePath string) {
	err := os.RemoveAll("output")
	if err != nil {
		panic(err)
	}
	cmd := exec.Command("java", "-jar", "libs/HypergraphTranslation.jar", "-convert", "-csp", filePath)
	if err = cmd.Run(); err != nil {
		panic(err)
	}
}

func HypertreeDecomposition(filePath string) {
	hypergraphPath := strings.ReplaceAll(filePath, ".xml", "hypergraph.hg")
	hypergraphPath = fmt.Sprintf("output/" + hypergraphPath)
	var name string
	switch runtime.GOOS {
	case "windows":
		name = "libs/balanced.exe"
	case "linux":
		name = "./libs/balancedLinux"
	}
	cmd := exec.Command(name, "-exact", "-graph", hypergraphPath, "-det", "-gml", "output/hypertree")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func GetHyperTree() (*Node, []*Node) {
	file, err := os.Open("hypertreeKakuro") //TODO: just debugging, change it
	if err != nil {
		panic(err)
	}
	nodes := make(map[int]*Node)
	var onlyNodes []*Node
	scanner := bufio.NewScanner(file)
	var line string
	for scanner.Scan() {
		line = scanner.Text()
		if strings.Contains(line, "node") {
			scanner.Scan()
			line = scanner.Text()
			reg := regexp.MustCompile("id (.*).*")
			res := reg.FindStringSubmatch(line)
			id, _ := strconv.Atoi(res[1])
			scanner.Scan()
			line = scanner.Text()
			reg = regexp.MustCompile("label \"{(.*)}\\s+{(.*)}\".*")
			res = reg.FindStringSubmatch(line)
			variables := strings.Split(res[2], ", ")
			node := Node{Id: id, Variables: variables}
			nodes[id] = &node
			onlyNodes = append(onlyNodes, &node)
		} else if strings.Contains(line, "edge") {
			scanner.Scan()
			line = scanner.Text()
			reg := regexp.MustCompile("source (.*).*")
			res := reg.FindStringSubmatch(line)
			source, _ := strconv.Atoi(res[1])
			scanner.Scan()
			line = scanner.Text()
			reg = regexp.MustCompile("target (.*).*")
			res = reg.FindStringSubmatch(line)
			target, _ := strconv.Atoi(res[1])
			nodes[source].AddSon(nodes[target])
		}
	}
	var root *Node
	for a := range nodes {
		if nodes[a].Father == nil {
			root = nodes[a]
			break
		}
	}
	err = file.Close()
	if err != nil {
		panic(err)
	}
	return root, onlyNodes
}

func GetConstraints(filePath string) []*Constraint {
	tablesPath := strings.ReplaceAll(filePath, ".xml", "tables.hg")
	tablesPath = fmt.Sprintf("output/" + tablesPath)
	file, err := os.Open(tablesPath)
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	var line string
	var constraints []*Constraint
	var c *Constraint
	phase1 := true
	phase2 := false
	phase3 := false
	scanner.Scan()
	line = scanner.Text()
	for {
		if phase1 {
			if line == "supports" {
				c = &Constraint{CType: true}
			} else if line == "conflicts" {
				c = &Constraint{CType: false}
			} else {
				panic("constraint " + line + " not supported")
			}
			phase2 = true
			phase1 = false
		} else if phase2 {
			scanner.Scan()
			line = scanner.Text()
			for _, v := range strings.Split(line, ",") {
				c.AddVariable(v)
			}
			phase2 = false
			phase3 = true
		} else if phase3 {
			for scanner.Scan() {
				line = scanner.Text()
				if line == "supports" || line == "conflicts" {
					phase3 = false
					phase1 = true
					break
				}
				possibleValuesString := strings.Split(line, ",")
				possibleValue := make([]int, 0)
				for _, s := range possibleValuesString {
					i, _ := strconv.Atoi(s)
					possibleValue = append(possibleValue, i)
				}
				c.AddPossibleValue(possibleValue)
			}
			constraints = append(constraints, c)
			if !phase1 {
				break
			}
		}
	}
	err = file.Close()
	if err != nil {
		panic(err)
	}
	return constraints
}

func GetDomains(filePath string) map[string][]int {
	domainPath := strings.ReplaceAll(filePath, ".xml", "domain.hg")
	domainPath = fmt.Sprintf("output/" + domainPath)
	file, err := os.Open(domainPath)
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	var line string
	m := make(map[string][]int)
	for scanner.Scan() {
		variable := scanner.Text()
		scanner.Scan()
		line = scanner.Text()
		values := make([]int, 0)
		for _, v := range strings.Split(line, " ") {
			i, _ := strconv.Atoi(v)
			values = append(values, i)
		}
		m[variable] = values
	}
	err = file.Close()
	if err != nil {
		panic(err)
	}
	return m
}
