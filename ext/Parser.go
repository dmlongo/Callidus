package ext

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/dmlongo/callidus/decomp"
)

// ParseDecomp from GML file
func ParseDecomp(htPath string) (*decomp.Node, []*decomp.Node) {
	file, err := os.Open(htPath)
	if err != nil {
		panic(err)
	}
	nodes := make(map[int]*decomp.Node)
	var onlyNodes []*decomp.Node
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
			node := decomp.Node{ID: id, Bag: variables, Lock: &sync.Mutex{}}
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
			nodes[source].AddChild(nodes[target])
		}
	}
	var root *decomp.Node
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

// ParseDecompInMemory from a string
func ParseDecompInMemory(htRaw *string) (*decomp.Node, []*decomp.Node) {
	output := strings.Split(*htRaw, "\n")
	nodes := make(map[int]*decomp.Node)
	var onlyNodes []*decomp.Node
	var fathersQueue []*decomp.Node
	idNodes := 0
	for i := 0; i < len(output); i++ {
		line := output[i]
		if strings.Contains(line, "Bag") {
			reg := regexp.MustCompile("Bag: {(.*)}.*")
			res := reg.FindStringSubmatch(line)
			variables := strings.Split(res[1], ", ")
			var nodeFather *decomp.Node = nil
			if len(fathersQueue) > 0 {
				nodeFather = fathersQueue[len(fathersQueue)-1]
			}
			node := decomp.Node{ID: idNodes, Bag: variables, Father: nodeFather, Lock: &sync.Mutex{}}
			if nodeFather != nil {
				nodeFather.AddChild(&node)
			}
			nodes[idNodes] = &node
			idNodes++
			onlyNodes = append(onlyNodes, &node)
		} else if strings.Contains(line, "Children") {
			fathersQueue = append(fathersQueue, nodes[idNodes-1])
		} else if strings.Contains(line, "]") {
			fathersQueue = fathersQueue[:len(fathersQueue)-1]
		}
	}
	var root *decomp.Node
	for a := range nodes {
		if nodes[a].Father == nil {
			root = nodes[a]
			break
		}
	}
	return root, onlyNodes
}

// ParseConstraints of a CSP
func ParseConstraints(cspPath string, folderName string) []*decomp.Constraint {
	var tablesPath string
	if strings.HasSuffix(cspPath, ".xml") {
		tablesPath = strings.ReplaceAll(cspPath, ".xml", "tables.hg")
	} else if strings.HasSuffix(cspPath, ".lzma") {
		tablesPath = strings.ReplaceAll(cspPath, ".lzma", "tables.hg")
	}
	tablesPath = fmt.Sprintf(folderName + tablesPath)
	file, err := os.Open(tablesPath)
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	var line string
	var constraints []*decomp.Constraint
	var c *decomp.Constraint

	const (
		readingCtype = iota
		readingVars
		readingTuples
		finished
	)
	phase := readingCtype
	scanner.Scan()
	line = scanner.Text()
	for phase != finished {
		switch phase {
		case readingCtype:
			if line == "supports" {
				c = &decomp.Constraint{CType: true}
			} else if line == "conflicts" {
				c = &decomp.Constraint{CType: false}
			} else {
				panic("constraint " + line + " not supported")
			}
			phase = readingVars
		case readingVars:
			scanner.Scan()
			line = scanner.Text()
			for _, v := range strings.Split(line, ",") {
				c.AddVariable(v)
			}
			phase = readingTuples
		case readingTuples:
			for scanner.Scan() {
				line = scanner.Text()
				if line == "supports" || line == "conflicts" {
					phase = readingCtype
					break
				} else {
					possibleValuesString := strings.Split(line, ",")
					possibleValue := make([]int, 0)
					for _, s := range possibleValuesString {
						i, _ := strconv.Atoi(s)
						possibleValue = append(possibleValue, i)
					}
					c.AddTuple(possibleValue)
				}
			}
			constraints = append(constraints, c)
			if phase != readingCtype {
				phase = finished
			}
		}
	}
	err = file.Close()
	if err != nil {
		panic(err)
	}
	return constraints
}

// ParseDomains of CSP variables
func ParseDomains(cspPath string, folder string) map[string][]int {
	var domainPath string
	if strings.HasSuffix(cspPath, ".xml") {
		domainPath = strings.ReplaceAll(cspPath, ".xml", "domain.hg")
	} else if strings.HasSuffix(cspPath, ".lzma") {
		domainPath = strings.ReplaceAll(cspPath, ".lzma", "domain.hg")
	}
	domainPath = fmt.Sprintf("output" + folder + domainPath)
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
