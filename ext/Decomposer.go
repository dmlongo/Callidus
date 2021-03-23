package ext

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var balancedGo string

func init() {
	path, err := os.Executable()
	if err != nil {
		panic(err)
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		panic(err)
	}
	balancedGo = filepath.Dir(path)
	switch runtime.GOOS {
	case "windows":
		balancedGo += "/libs/balanced.exe"
	case "linux":
		balancedGo += "/libs/BalancedGo"
	}
}

// Decompose the hypergraph of a CSP
func Decompose(hgPath string) string {
	// TODO add logging, check if you get errors if command is wrong
	out, err := exec.Command(balancedGo, "-graph", hgPath, "-approx", "3600", "-det").Output()
	if err != nil {
		panic(err)
	}
	return string(out)
}

// DecomposeToFile decompose a hypergraph and saves the decomposition on a file
func DecomposeToFile(hgPath string, htPath string) string {
	// TODO add logging, check if you get errors if command is wrong
	out, err := exec.Command(balancedGo, "-graph", hgPath, "-approx", "3600", "-det", "-gml", htPath).Output()
	if err != nil {
		panic(err)
	}
	return string(out)
}
