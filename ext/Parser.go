package ext

import (
	"bufio"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/dmlongo/callidus/ctr"
	"github.com/dmlongo/callidus/decomp"
	files "github.com/dmlongo/callidus/ext/files"
)

var regexNodeID = regexp.MustCompile(`id (.*).*`)
var regexNodeLabel = regexp.MustCompile(`label "{(.*)}\s+{(.*)}".*`)
var regexEdgeSource = regexp.MustCompile(`source (.*).*`)
var regexEdgeTarget = regexp.MustCompile(`target (.*).*`)

var regexBag = regexp.MustCompile(`Bag: {(.*)}.*`)
var regexCover = regexp.MustCompile(`Cover: {(.*)}.*`)

// ParseDecompFromFile from GML file
func ParseDecompFromFile(htPath string) (*decomp.Node, []*decomp.Node) {
	file, err := os.Open(htPath)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	nodes := make(map[int]*decomp.Node)
	var onlyNodes []*decomp.Node
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
			node := decomp.NewNode(id, variables, edges)
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

	var root *decomp.Node
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
			res := regexBag.FindStringSubmatch(line)
			if len(res) < 2 {
				panic("Cannot parse bag in: " + line)
			}
			variables := strings.Split(res[1], ",")
			trimSpaces(variables)
			sort.Strings(variables)
			var nodeFather *decomp.Node = nil
			if len(fathersQueue) > 0 {
				nodeFather = fathersQueue[len(fathersQueue)-1]
			}
			node := decomp.NewNode(idNodes, variables, nil)
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
	var root *decomp.Node
	for a := range nodes {
		if nodes[a].Parent == nil {
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
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	reader := bufio.NewReader(file)
	constraints := make(map[string]ctr.Constraint)
	numLines := 0
	for {
		line, eof := files.ReadLineCount(reader, &numLines)
		if eof {
			break
		}

		var name string
		var constr ctr.Constraint
		switch line {
		case "ExtensionCtr":
			name, _ = files.ReadLineCount(reader, &numLines)
			vars, _ := files.ReadLineCount(reader, &numLines)
			ctype, _ := files.ReadLineCount(reader, &numLines)
			tuples, _ := files.ReadLineCount(reader, &numLines)
			constr = &ctr.ExtensionCtr{CName: name, Vars: vars, CType: ctype, Tuples: tuples}
		case "PrimitiveCtr":
			name, _ = files.ReadLineCount(reader, &numLines)
			vars, _ := files.ReadLineCount(reader, &numLines)
			f, _ := files.ReadLineCount(reader, &numLines)
			constr = &ctr.PrimitiveCtr{CName: name, Vars: vars, Function: f}
		case "AllDifferentCtr":
			name, _ = files.ReadLineCount(reader, &numLines)
			vars, _ := files.ReadLineCount(reader, &numLines)
			constr = &ctr.AllDifferentCtr{CName: name, Vars: vars}
		case "ElementCtr":
			name, _ = files.ReadLineCount(reader, &numLines)
			vars, _ := files.ReadLineCount(reader, &numLines)
			list, _ := files.ReadLineCount(reader, &numLines)
			startIndex, _ := files.ReadLineCount(reader, &numLines)
			index, _ := files.ReadLineCount(reader, &numLines)
			rank, _ := files.ReadLineCount(reader, &numLines)
			condition, _ := files.ReadLineCount(reader, &numLines)
			constr = &ctr.ElementCtr{CName: name, Vars: vars, List: list, StartIndex: startIndex, Index: index, Rank: rank, Condition: condition}
		case "SumCtr":
			name, _ = files.ReadLineCount(reader, &numLines)
			vars, _ := files.ReadLineCount(reader, &numLines)
			coeffs, _ := files.ReadLineCount(reader, &numLines)
			condition, _ := files.ReadLineCount(reader, &numLines)
			constr = &ctr.SumCtr{CName: name, Vars: vars, Coeffs: coeffs, Condition: condition}
		default:
			msg := ctrFile + ", line " + strconv.Itoa(numLines) + ": " + line + " not implemented yet"
			panic(msg)
		}
		constraints[name] = constr
	}

	return constraints
}

// ParseDomains of CSP variables
func ParseDomains(domFile string) map[string]string {
	file, err := os.Open(domFile)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	reader := bufio.NewReader(file)
	m := make(map[string]string)
	for {
		line, eof := files.ReadLine(reader)
		if eof {
			break
		}

		tks := strings.Split(line, ";")
		variable := tks[0]
		domain := tks[1]
		m[variable] = domain
	}

	return m
}
