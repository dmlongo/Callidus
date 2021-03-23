package ext

import (
	"os"
	"os/exec"

	"github.com/dmlongo/callidus/decomp"
)

// Convert a CSP into a hypergraph
func Convert(cspPath string, outDir string) decomp.Hypergraph {
	// TODO add logging
	execPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	execPath += "/libs/hgtools.jar"
	cmd := exec.Command("java", "-jar", execPath, "-convert", "-xcsp", "-print", "-out", outDir, cspPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	return decomp.BuildHypergraph(stdout)
}
