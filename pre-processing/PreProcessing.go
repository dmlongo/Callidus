package pre_processing

import (
	. "../../Callidus/constraint"
	. "../../Callidus/hyperTree"
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

func HypergraphTranslation(filePath string) {
	err := os.RemoveAll("output")
	if err != nil {
		panic(err)
	}
	cmd := exec.Command("java", "-jar", "libs/HypergraphTranslation.jar", "-convert", "-csp", filePath)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		panic(fmt.Sprint(err) + ": " + stderr.String())
	}
}

func HypertreeDecomposition(filePath string, folderName string, algorithm string, inMemory bool, computeWidth bool) string {
	var hypergraphPath string
	if strings.HasSuffix(filePath, ".xml") {
		hypergraphPath = strings.ReplaceAll(filePath, ".xml", "hypergraph.hg")
	} else if strings.HasSuffix(filePath, ".lzma") {
		hypergraphPath = strings.ReplaceAll(filePath, ".lzma", "hypergraph.hg")
	}
	hypergraphPath = fmt.Sprintf(folderName + hypergraphPath)

	var name string
	switch runtime.GOOS {
	case "windows":
		name = "libs/balanced.exe"
	case "linux":
		name = "./libs/balancedLinux"
	}

	//TODO: we must find another way to write it
	var width string
	if computeWidth {
		cmd := exec.Command("python3", "libs/widthComputation.py", hypergraphPath)
		byte, err := cmd.Output()
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err != nil {
			panic(fmt.Sprint(err) + ": " + stderr.String())
		}
		width = strings.ReplaceAll(string(byte), "\n", "")
	}
	var cmd *exec.Cmd
	if inMemory {
		if computeWidth {
			cmd = exec.Command(name, "-width", width, "-graph", hypergraphPath, "-det")
		} else {
			if algorithm == "det" {
				cmd = exec.Command(name, "-exact", "-graph", hypergraphPath, "-det")
			} else if algorithm == "balDet" {
				cmd = exec.Command(name, "-exact", "-graph", hypergraphPath, "-balDet", "1")
			}
		}
		byte, err := cmd.Output()
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err != nil {
			panic(fmt.Sprint(err) + ": " + stderr.String())
		}
		return string(byte)
	} else {
		if computeWidth {
			cmd = exec.Command(name, "-width", width, "-graph", hypergraphPath, "-det", "-gml", folderName+"hypertree")
		} else {
			if algorithm == "det" {
				cmd = exec.Command(name, "-exact", "-graph", hypergraphPath, "-det", "-gml", folderName+"hypertree")
			} else if algorithm == "balDet" {
				cmd = exec.Command(name, "-exact", "-graph", hypergraphPath, "-balDet", "1", "-gml", folderName+"hypertree")
			}
		}
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			panic(fmt.Sprint(err) + ": " + stderr.String())
		}
		return ""
	}
}

func GetHyperTree(filePath string) (*Node, []*Node) {
	file, err := os.Open(filePath)
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
			node := Node{Id: id, Variables: variables, Lock: &sync.Mutex{}}
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

func GetHyperTreeInMemory(hyperTreeRaw *string) (*Node, []*Node) {
	output := strings.Split(*hyperTreeRaw, "\n")
	nodes := make(map[int]*Node)
	var onlyNodes []*Node
	var fathersQueue []*Node
	idNodes := 0
	for i := 0; i < len(output); i++ {
		line := output[i]
		if strings.Contains(line, "Bag") {
			reg := regexp.MustCompile("Bag: {(.*)}.*")
			res := reg.FindStringSubmatch(line)
			variables := strings.Split(res[1], ", ")
			var nodeFather *Node = nil
			if len(fathersQueue) > 0 {
				nodeFather = fathersQueue[len(fathersQueue)-1]
			}
			node := Node{Id: idNodes, Variables: variables, Father: nodeFather, Lock: &sync.Mutex{}}
			if nodeFather != nil {
				nodeFather.AddSon(&node)
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
	var root *Node
	for a := range nodes {
		if nodes[a].Father == nil {
			root = nodes[a]
			break
		}
	}
	return root, onlyNodes
}

func GetConstraints(filePath string, folderName string) []*Constraint {
	var tablesPath string
	if strings.HasSuffix(filePath, ".xml") {
		tablesPath = strings.ReplaceAll(filePath, ".xml", "tables.hg")
	} else if strings.HasSuffix(filePath, ".lzma") {
		tablesPath = strings.ReplaceAll(filePath, ".lzma", "tables.hg")
	}
	tablesPath = fmt.Sprintf(folderName + tablesPath)
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

func GetDomains(filePath string, folderName string) map[string][]int {
	var domainPath string
	if strings.HasSuffix(filePath, ".xml") {
		domainPath = strings.ReplaceAll(filePath, ".xml", "domain.hg")
	} else if strings.HasSuffix(filePath, ".lzma") {
		domainPath = strings.ReplaceAll(filePath, ".lzma", "domain.hg")
	}
	domainPath = fmt.Sprintf(folderName + domainPath)
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
