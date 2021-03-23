package ext

import (
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
		panic(err)
	}

	output := string(out)
	lines := strings.Split(output, "\n")
	if strings.HasPrefix(lines[2], "OK") {
		return output, true
	}
	return output, false
}
