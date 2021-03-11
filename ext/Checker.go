package ext

import (
	"os/exec"
	"strings"

	"github.com/dmlongo/callidus/ctr"
)

// CheckSolution of a CSP
func CheckSolution(csp string, solution map[string]int) (string, bool) {
	xcspSol := ctr.WriteSolution(solution)
	out, err := exec.Command("java", "-cp", "libs/xcsp3-tools-1.2.3.jar", "org.xcsp.parser.callbacks.SolutionChecker", csp, xcspSol).Output()
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
