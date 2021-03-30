package ext

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dmlongo/callidus/decomp"
)

var hgtools string

func init() {
	path, err := os.Executable()
	if err != nil {
		panic(err)
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		panic(err)
	}
	hgtools = filepath.Dir(path) + "/libs/hgtools.jar"
}

// Convert a CSP into a hypergraph
func Convert(cspPath string, outDir string) decomp.Hypergraph {
	// TODO add logging
	fmt.Println("hgtools=", hgtools)
	cmd := exec.Command("java", "-jar", hgtools, "-convert", "-xcsp", "-print", "-out", outDir, cspPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	reader := bufio.NewReader(stdout)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF && len(line) == 0 {
			break
		}
		fmt.Println("line=", line)
	}
	return nil

	//return decomp.BuildHypergraph(stdout)
}
