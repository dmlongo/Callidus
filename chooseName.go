package main

import (
	. "../CSP_Project/constraint"
	. "../CSP_Project/hyperTree"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

func main() {
	//TODO: Aggiornare gitignore
	filePath := "3col.xml"
	hypergraphTranslation(filePath)
	hypertreeDecomposition(filePath)
	_, nodes := getHyperTree() //TODO: NON FATE GLI STRONZI
	var wg sync.WaitGroup
	wg.Add(2)
	var domains map[string][]int
	go func() {
		domains = getDomains(filePath)
		wg.Done()
	}()
	var constraints []*Constraint
	go func() {
		constraints = getConstraints(filePath) //TODO: parallelizzabile?
		wg.Done()
	}()
	wg.Wait()
	createAndSolveSubCSP(domains, constraints, nodes)
}

func hypergraphTranslation(filePath string) {
	cmd := exec.Command("java", "-jar", "libs/HypergraphTranslation.jar", "-convert", "-csp", filePath)
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func hypertreeDecomposition(filePath string) {
	hypergraphPath := strings.ReplaceAll(filePath, ".xml", "hypergraph.hg")
	hypergraphPath = fmt.Sprintf("output/" + hypergraphPath)
	//check width
	cmd := exec.Command("./libs/balanced.exe", "-exact", "-graph", hypergraphPath, "-det", "-gml", "output/hypertree")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func getHyperTree() (*Node, []*Node) {
	file, err := os.Open("output/hypertree")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	nodes := make(map[int]*Node)
	var onlyNodes []*Node
	scanner := bufio.NewScanner(file)
	var line string
	for scanner.Scan() {
		line = scanner.Text()
		if strings.Contains(line, "node") {
			scanner.Scan() //TODO assert?
			line = scanner.Text()
			reg := regexp.MustCompile("id (.*).*")
			res := reg.FindStringSubmatch(line)
			id, _ := strconv.Atoi(res[1])
			scanner.Scan() //TODO assert?
			line = scanner.Text()
			reg = regexp.MustCompile("label \"{(.*)} {(.*)}\".*")
			res = reg.FindStringSubmatch(line)
			joinNodes := strings.Split(res[1], ", ")
			variables := strings.Split(res[2], ", ")
			node := Node{Id: id, JoinNodes: joinNodes, Variables: variables}
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
			nodes[target].AddFather(nodes[source])
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

func getConstraints(filePath string) []*Constraint {
	tablesPath := strings.ReplaceAll(filePath, ".xml", "tables.hg")
	tablesPath = fmt.Sprintf("output/" + tablesPath)
	file, err := os.Open(tablesPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
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
	return constraints
}

func getDomains(filePath string) map[string][]int {
	domainPath := strings.ReplaceAll(filePath, ".xml", "domain.hg")
	domainPath = fmt.Sprintf("output/" + domainPath)
	file, err := os.Open(domainPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
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
	return m
}

//TODO: provare a parallelizzare la scrittura su file
func createAndSolveSubCSP(domains map[string][]int, constraints []*Constraint, nodes []*Node) {
	os.RemoveAll("subCSP")
	err := os.Mkdir("subCSP", 777)
	if err != nil {
		panic(err)
	}
	for _, node := range nodes {
		fileName := "subCSP/" + strconv.Itoa(node.Id) + ".xml"
		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND, 777)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		file.WriteString("<instance format=\"XCSP3\" type=\"CSP\">\n")
		file.WriteString("<variables>\n")
		for _, variable := range node.Variables {
			dom := domains[variable]
			values := "<var id=\"" + variable + "\"> "
			for _, i := range dom {
				values += strconv.Itoa(i) + " "
			}
			values += "</var>\n"
			file.WriteString(values)
		}
		file.WriteString("</variables>\n")
		file.WriteString("<constraints>\n")
		for _, constraint := range constraints {
			isConstraintOk := true
			for _, constraintVariable := range constraint.Variables {
				isConstraintVariableOk := false
				for _, nodeVariable := range node.Variables {
					if constraintVariable == nodeVariable {
						isConstraintVariableOk = true
						break
					}
				}
				if !isConstraintVariableOk {
					isConstraintOk = false
				}
			}
			if isConstraintOk {
				file.WriteString("<extension>\n")
				listVariable := "<list> "
				for _, v := range constraint.Variables {
					listVariable += v + " "
				}
				listVariable += "</list>\n"
				file.WriteString(listVariable)
				possibleValues := "<supports> "
				if !constraint.CType {
					possibleValues = "<conflicts> "
				}
				for _, tup := range constraint.PossibleValues {
					possibleValues += "("
					for i := range tup {
						value := strconv.Itoa(tup[i])
						if i == len(tup)-1 {
							possibleValues += value
						} else {
							possibleValues += value + ","
						}
					}
					possibleValues += ")"
				}
				if constraint.CType {
					possibleValues += " </supports>\n"
				} else {
					possibleValues += " </conflicts>\n"
				}
				file.WriteString(possibleValues)
				file.WriteString("</extension>\n")
			}
		}
		file.WriteString("</constraints>\n")
		file.WriteString("</instance>\n")
		file.Close()
		outputFileName := strings.ReplaceAll(fileName, ".xml", "sol.txt")
		outputFile, err := os.OpenFile(outputFileName, os.O_CREATE|os.O_APPEND, 777)
		if err != nil {
			panic(err)
		}
		out, err := exec.Command("java", "-cp", "libs/AbsCon.jar", "AbsCon", fileName, "-s=all").Output()
		if err != nil {
			panic(err)
		}
		outputFile.Write(out)
	}
}
