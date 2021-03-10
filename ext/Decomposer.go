package ext

import (
	"os/exec"
	"runtime"
)

// Decompose the hypergraph of a CSP
func Decompose(hgPath string) string {
	var cmdName string
	switch runtime.GOOS {
	case "windows":
		cmdName = "libs/balanced.exe"
	case "linux":
		cmdName = "./libs/BalancedGo"
	}

	// TODO add logging, check if you get errors if command is wrong
	out, err := exec.Command(cmdName, "-graph", hgPath, "-approx", "3600", "-det").Output()
	if err != nil {
		panic(err)
	}
	return string(out)
}

// DecomposeToFile decompose a hypergraph and saves the decomposition on a file
func DecomposeToFile(hgPath string, htPath string) string {
	var name string
	switch runtime.GOOS {
	case "windows":
		name = "libs/balanced.exe"
	case "linux":
		name = "./libs/BalancedGo"
	}

	// TODO add logging, check if you get errors if command is wrong
	out, err := exec.Command(name, "-graph", hgPath, "-approx", "3600", "-det", "-gml", htPath).Output()
	if err != nil {
		panic(err)
	}
	return string(out)
}
