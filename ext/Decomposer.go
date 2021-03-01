package ext

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Decompose the hypergraph of a CSP
func Decompose(cspPath string, folder string, inMemory bool) string {
	var hypergraphPath string
	if strings.HasSuffix(cspPath, ".xml") {
		hypergraphPath = strings.ReplaceAll(cspPath, ".xml", "hypergraph.hg")
	} else if strings.HasSuffix(cspPath, ".lzma") {
		hypergraphPath = strings.ReplaceAll(cspPath, ".lzma", "hypergraph.hg")
	}
	hypergraphPath = fmt.Sprintf(folder + hypergraphPath)

	var name string
	switch runtime.GOOS {
	case "windows":
		name = "libs/balanced.exe"
	case "linux":
		name = "./libs/balancedLinux"
	}

	// TODO add logging, avoid code repetition, check if you get errors if command is wrong
	var cmd *exec.Cmd
	if inMemory {
		cmd = exec.Command(name, "-graph", hypergraphPath, "-approx", "3600", "-det")
		output, err := cmd.Output()
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err != nil {
			panic(fmt.Sprint(err) + ": " + stderr.String())
		}
		return string(output)
	}
	cmd = exec.Command(name, "-graph", hypergraphPath, "-approx", "3600", "-det", "-gml", folder+"hypertree")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		panic(fmt.Sprint(err) + ": " + stderr.String())
	}
	return ""
}
