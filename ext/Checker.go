package ext

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/dmlongo/callidus/ctr"
)

// CheckSolution of a CSP
func CheckSolution(csp string, solution ctr.Solution) (string, bool) {
	xcspSol := ctr.WriteSolution(solution)
	execPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	execPath += "/libs/xcsp3-tools-1.2.3.jar"
	fmt.Println("execPath=", execPath)
	out, err := exec.Command("java", "-cp", execPath, "org.xcsp.parser.callbacks.SolutionChecker", csp, xcspSol).Output()
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
