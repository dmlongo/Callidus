package ext

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
func Decompose(hgPath string, timeout string) string {
	// TODO add logging
	out, err := exec.Command(balancedGo, "-graph", hgPath, "-approx", timeout, "-det", "-bench").Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			panic(fmt.Sprintf("BalancedGo failed: %v: %s", err, ee.Stderr))
		} else {
			panic(fmt.Sprintf("BalancedGo failed: %v", err))
		}
	}
	res := string(out)
	if strings.HasSuffix(res, "false\n") {
		return ""
	}
	return res
}

// DecomposeToFile decompose a hypergraph and saves the decomposition on a file
func DecomposeToFile(hgPath string, htPath string, timeout string) string {
	// TODO add logging
	out, err := exec.Command(balancedGo, "-graph", hgPath, "-approx", timeout, "-det", "-gml", htPath, "-bench").Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			panic(fmt.Sprintf("BalancedGo failed: %v: %s", err, ee.Stderr))
		} else {
			panic(fmt.Sprintf("BalancedGo failed: %v", err))
		}
	}
	res := string(out)
	if strings.HasSuffix(res, "false\n") {
		return ""
	}
	return res
}
