package ext

import (
	"os"
	"os/exec"

	"github.com/dmlongo/callidus/decomp"
)

// Convert a CSP into a hypergraph
func Convert(cspPath string, outDir string) decomp.Hypergraph {
	// TODO add logging
	err := os.RemoveAll(outDir) // TODO removing wastes time, not necessary
	if err != nil {
		panic(err)
	}
	cmd := exec.Command("java", "-jar", "libs/hgtools.jar", "-convert", "-xcsp", "-print", "-out", outDir, cspPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	return decomp.BuildHypergraph(stdout)
}
