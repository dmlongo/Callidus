package ext

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dmlongo/callidus/ctr"
)

var xcsp3Tools string

func init() {
	path, err := os.Executable()
	if err != nil {
		panic(err)
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		panic(err)
	}
	xcsp3Tools = filepath.Dir(path) + "/libs/xcsp3-tools-1.2.3.jar"
}

// CheckSolution of a CSP
func CheckSolution(csp string, solution ctr.Solution) (string, bool) {
	xcspSol := ctr.WriteSolution(solution)
	out, err := exec.Command("java", "-cp", xcsp3Tools, "org.xcsp.parser.callbacks.SolutionChecker", csp, xcspSol).Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			panic(fmt.Sprintf("xcsp3-tools failed: %v: %s", err, ee.Stderr))
		} else {
			panic(fmt.Sprintf("xcsp3-tools failed: %v", err))
		}
	}

	output := string(out)
	lines := strings.Split(output, "\n")
	if strings.HasPrefix(lines[2], "OK") {
		return output, true
	}
	return output, false
}
