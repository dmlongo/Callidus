package main

import (
	. "../CSP_Project/hyperTree"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	filePath := "3col.xml"
	hypergraphTranslation(filePath)
	hypertreeDecomposition(filePath)
	getHyperTree()
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

func getHyperTree() {
	file, err := os.Open("output/hypertree")
	if err != nil {
		panic(err)
	}
	nodes := make(map[int]*Node)
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
	for a := range nodes {
		fmt.Println(nodes[a])
	}
}
