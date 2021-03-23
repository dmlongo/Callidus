package ext

import (
	"os"
	"os/exec"
	"runtime"
)

// Decompose the hypergraph of a CSP
func Decompose(hgPath string) string {
	execPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	switch runtime.GOOS {
	case "windows":
		execPath += "/libs/balanced.exe"
	case "linux":
		execPath += "/libs/BalancedGo"
	}

	// TODO add logging, check if you get errors if command is wrong
	out, err := exec.Command(execPath, "-graph", hgPath, "-approx", "3600", "-det").Output()
	if err != nil {
		panic(err)
	}
	return string(out)
}

// DecomposeToFile decompose a hypergraph and saves the decomposition on a file
func DecomposeToFile(hgPath string, htPath string) string {
	execPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	switch runtime.GOOS {
	case "windows":
		execPath += "/libs/balanced.exe"
	case "linux":
		execPath += "/libs/BalancedGo"
	}

	// TODO add logging, check if you get errors if command is wrong
	out, err := exec.Command(execPath, "-graph", hgPath, "-approx", "3600", "-det", "-gml", htPath).Output()
	if err != nil {
		panic(err)
	}
	return string(out)
}
