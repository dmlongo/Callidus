package ext

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// Convert a CSP into a hypergraph
func Convert(cspPath string) {
	// TODO add logging
	err := os.RemoveAll("output")
	if err != nil {
		panic(err)
	}
	cmd := exec.Command("java", "-jar", "libs/hgtools.jar", "-convert", "-xcsp", cspPath)
	//var out bytes.Buffer
	var stderr bytes.Buffer
	//cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil { // TODO doesn't throw error when command is wrong
		panic(fmt.Sprint(err) + ":" + stderr.String())
	}
}
