package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func main() {
	filePath := "input/3col.xml"
	hypergraphTranslation(filePath)
	hypertreeDecomposition(filePath)
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
