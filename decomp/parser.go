package decomp

import (
	"bufio"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/dmlongo/callidus/files"
)

var regexNodeID = regexp.MustCompile(`id (.*).*`)
var regexNodeLabel = regexp.MustCompile(`label "{(.*)}\s+{(.*)}".*`)
var regexEdgeSource = regexp.MustCompile(`source (.*).*`)
var regexEdgeTarget = regexp.MustCompile(`target (.*).*`)

var regexBag = regexp.MustCompile(`Bag: {(.*)}.*`)
var regexCover = regexp.MustCompile(`Cover: {(.*)}.*`)

// ParseGML parses a decomposition file in GML format
func ParseGML(htPath string) (*Node, []*Node) {
	file, err := os.Open(htPath)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	nodes := make(map[int]*Node)
	var onlyNodes []*Node
	reader := bufio.NewReader(file)
	numLines := 0
	for {
		line, eof := files.ReadLineCount(reader, &numLines)
		if eof {
			break
		}

		if strings.Contains(line, "node") {
			line, _ = files.ReadLineCount(reader, &numLines)
			res := regexNodeID.FindStringSubmatch(line)
			if len(res) < 2 {
				panic("Cannot parse node ID in line" + strconv.Itoa(numLines) + ": " + line)
			}
			id, _ := strconv.Atoi(res[1])
			line, _ = files.ReadLineCount(reader, &numLines)
			res = regexNodeLabel.FindStringSubmatch(line)
			if len(res) < 3 {
				panic("Cannot parse node label in line" + strconv.Itoa(numLines) + ": " + line)
			}
			edges := strings.Split(res[1], ",")
			trimSpaces(edges)
			variables := strings.Split(res[2], ",")
			trimSpaces(variables)
			sort.Strings(variables)
			node := NewNode(id, variables, edges)
			nodes[id] = node
			onlyNodes = append(onlyNodes, node)
		} else if strings.Contains(line, "edge") {
			line, _ = files.ReadLineCount(reader, &numLines)
			res := regexEdgeSource.FindStringSubmatch(line)
			if len(res) < 2 {
				panic("Cannot parse edge source in line" + strconv.Itoa(numLines) + ": " + line)
			}
			source, _ := strconv.Atoi(res[1])
			line, _ = files.ReadLineCount(reader, &numLines)
			res = regexEdgeTarget.FindStringSubmatch(line)
			if len(res) < 2 {
				panic("Cannot parse edge target in line" + strconv.Itoa(numLines) + ": " + line)
			}
			target, _ := strconv.Atoi(res[1])
			nodes[source].AddChild(nodes[target])
		}
	}

	var root *Node
	for a := range nodes {
		if nodes[a].Parent == nil {
			root = nodes[a]
			break
		}
	}

	return root, onlyNodes
}

func trimSpaces(arr []string) {
	for i := 0; i < len(arr); i++ {
		arr[i] = strings.TrimSpace(arr[i])
	}
}

// ParseBalancedGo parses a decomposition in the BalancedGo output format
func ParseBalancedGo(htRaw *string) (*Node, []*Node) {
	output := strings.Split(*htRaw, "\n")
	nodes := make(map[int]*Node)
	var onlyNodes []*Node
	var fathersQueue []*Node
	idNodes := 0
	for i := 0; i < len(output); i++ {
		line := output[i]
		if strings.Contains(line, "Bag") {
			res := regexBag.FindStringSubmatch(line)
			if len(res) < 2 {
				panic("Cannot parse bag in: " + line)
			}
			variables := strings.Split(res[1], ",")
			trimSpaces(variables)
			sort.Strings(variables)
			var nodeFather *Node = nil
			if len(fathersQueue) > 0 {
				nodeFather = fathersQueue[len(fathersQueue)-1]
			}
			node := NewNode(idNodes, variables, nil)
			if nodeFather != nil {
				nodeFather.AddChild(node)
			}
			nodes[idNodes] = node
			idNodes++
			onlyNodes = append(onlyNodes, node)
		} else if strings.Contains(line, "Cover") {
			res := regexCover.FindStringSubmatch(line)
			if len(res) < 2 {
				panic("Cannot parse cover in: " + line)
			}
			cov := strings.Split(res[1], ",")
			trimSpaces(cov)
			n := onlyNodes[idNodes-1]
			n.SetCover(cov)
		} else if strings.Contains(line, "Children") {
			fathersQueue = append(fathersQueue, nodes[idNodes-1])
		} else if strings.Contains(line, "]") {
			fathersQueue = fathersQueue[:len(fathersQueue)-1]
		}
	}
	var root *Node
	for a := range nodes {
		if nodes[a].Parent == nil {
			root = nodes[a]
			break
		}
	}
	return root, onlyNodes
}
