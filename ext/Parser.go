package ext

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/dmlongo/callidus/ctr"
	"github.com/dmlongo/callidus/decomp"
)

// ParseDecompFromFile from GML file
func ParseDecompFromFile(htPath string) (*decomp.Node, []*decomp.Node) {
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
			edges := strings.Split(res[1], ", ")
			variables := strings.Split(res[2], ", ")
			node := decomp.Node{ID: id, Lock: &sync.Mutex{}}
			node.SetBag(variables)
			node.SetCover(edges)
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

// ParseDecomp from a string
func ParseDecomp(htRaw *string) (*decomp.Node, []*decomp.Node) {
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
			node := decomp.Node{ID: idNodes, Father: nodeFather, Lock: &sync.Mutex{}}
			node.SetBag(variables)
			if nodeFather != nil {
				nodeFather.AddChild(&node)
			}
			nodes[idNodes] = &node
			idNodes++
			onlyNodes = append(onlyNodes, &node)
		} else if strings.Contains(line, "Cover") {
			reg := regexp.MustCompile("Cover: {(.*)}.*")
			res := reg.FindStringSubmatch(line)
			cov := strings.Split(res[1], ", ")
			n := onlyNodes[idNodes-1]
			n.SetCover(cov)
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
func ParseConstraints(ctrFile string) map[string]ctr.Constraint {
	file, err := os.Open(ctrFile)
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	constraints := make(map[string]ctr.Constraint)
	for scanner.Scan() {
		var name string
		var constr ctr.Constraint
		switch line := scanner.Text(); line {
		case "ExtensionCtr":
			scanner.Scan()
			name = scanner.Text()
			scanner.Scan()
			vars := scanner.Text()
			scanner.Scan()
			ctype := scanner.Text()
			scanner.Scan()
			tuples := scanner.Text()
			constr = &ctr.ExtensionCtr{CName: name, Vars: vars, CType: ctype, Tuples: tuples}
		default:
			fmt.Println(line + " not implemented yet")
		}
		constraints[name] = constr
	}
	err = file.Close()
	if err != nil {
		panic(err)
	}
	return constraints
}

// ParseDomains of CSP variables
func ParseDomains(domFile string) map[string]string {
	file, err := os.Open(domFile)
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	var line string
	m := make(map[string]string)
	for scanner.Scan() {
		line = scanner.Text()
		tks := strings.Split(line, ";")
		variable := tks[0]
		domain := tks[1]
		m[variable] = domain
	}
	err = file.Close()
	if err != nil {
		panic(err)
	}
	return m
}
