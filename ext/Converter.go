package ext

import (
	"bufio"
	"bytes"
	"fmt"
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
	cmd := exec.Command("java", "-jar", hgtools, "-convert", "-xcsp", "-print", "-out", outDir, cspPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	hg := decomp.BuildHypergraph(bufio.NewReader(stdout))
	if err := cmd.Wait(); err != nil {
		panic(fmt.Sprintf("hgtools failed: %v: %s", err, stderr.String()))
	}
	return hg
}
