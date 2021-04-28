package decomp

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/dmlongo/callidus/csp"
	"github.com/dmlongo/callidus/db"
	files "github.com/dmlongo/callidus/ext/files"
)

var nacre string

func init() {
	path, err := os.Executable()
	if err != nil {
		panic(err)
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		panic(err)
	}
	nacre = filepath.Dir(path) + "/libs/nacre"
}

var listRegex = regexp.MustCompile(`.*<list>(.*)</list>.*`)
var valuesRegex = regexp.MustCompile(`.*<values>(.*)</values>.*`)

// SolveSubCspSeq solve the CSPs associated to a hypertree sequentially
func SolveSubCspSeq(nodes []*Node, domains map[string]string, constraints map[string]csp.Constraint, baseDir string) bool {
	subCspFolder := files.MakeDir(baseDir + "subs/")

	sat := true
	for _, node := range nodes {
		nodeCtrs, nodeVars := filterCtrsVars(node, constraints, domains)
		subFile := subCspFolder + "sub" + strconv.Itoa(node.ID) + ".xml"
		csp.CreateXCSPInstance(nodeCtrs, nodeVars, subFile)
		sat = solveCSPSeq(subFile, node)
		if !sat {
			break
		}
	}
	return sat
}

func solveCSPSeq(cspFile string, node *Node) bool {
	cmd := exec.Command(nacre, cspFile, "-complete", "-sols", "-verb=3")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	res := readTuples(bufio.NewReader(stdout), cspFile, node)
	if err := cmd.Wait(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() != 40 {
			panic(fmt.Sprintf("nacre failed: %v:", err))
		}
	}
	return res
}

func readTuples(reader *bufio.Reader, cspFile string, node *Node) bool {
	solFound := false
	for {
		line, eof := files.ReadLine(reader)
		if eof {
			break
		}
		if strings.HasPrefix(line, "v") {
			tup := makeTuple(line, cspFile, node.bagSet)
			if _, added := node.Tuples.AddTuple(tup); !added {
				panic(fmt.Sprintf("node %v, %s: Could not add tuple %v", node.ID, cspFile, tup))
			}
			solFound = true
		}
	}
	return solFound
}

func makeTuple(line string, cspFile string, bag map[string]int) db.Tuple {
	matchesVal := valuesRegex.FindStringSubmatch(line)
	if len(matchesVal) < 2 {
		panic(cspFile + ", bad values= " + line)
	}
	matchesList := listRegex.FindStringSubmatch(line)
	if len(matchesList) < 2 {
		panic(cspFile + ", bad list= " + line)
	}
	list := strings.Split(strings.TrimSpace(matchesList[1]), " ")
	tup := make([]int, len(bag))
	z := 0
	for i, value := range strings.Split(strings.TrimSpace(matchesVal[1]), " ") {
		v, err := strconv.Atoi(value)
		if err != nil {
			panic(err)
		}
		if _, ok := bag[list[i]]; ok {
			tup[z] = v
			z++
		}
	}
	if z != len(bag) {
		panic(fmt.Sprintf("Did not find enough variables %v/%v, list: %v", z, len(bag), list))
	}
	return tup
}

// SolveSubCspPar solve the CSPs associated to a hypertree in parallel
func SolveSubCspPar(nodes []*Node, domains map[string]string, constraints map[string]csp.Constraint, baseDir string) bool {
	subCspFolder := files.MakeDir(baseDir + "subs/")

	jobs := make(chan *Node)
	go func() {
		for _, node := range nodes {
			jobs <- node
		}
		close(jobs)
	}()

	sat := make(chan bool)
	quit := make(chan bool)
	defer close(quit)
	numNodes := len(nodes)
	numWorkers := runtime.NumCPU()
	var wg sync.WaitGroup
	wg.Add(numNodes)
	if numNodes < numWorkers {
		numWorkers = numNodes
	}
	for i := 0; i < numWorkers; i++ {
		go func() { // launch a worker
			for n := range jobs {
				nodeCtrs, nodeVars := filterCtrsVars(n, constraints, domains)
				subFile := subCspFolder + "sub" + strconv.Itoa(n.ID) + ".xml"
				csp.CreateXCSPInstance(nodeCtrs, nodeVars, subFile)
				solveCSPPar(subFile, n, sat, quit)
				wg.Done()
			}
		}()
	}
	go func() {
		wg.Wait()
		close(sat)
	}()

	for i := 0; i < len(nodes); i++ {
		if !<-sat {
			return false
		}
	}
	return true
}

func solveCSPPar(cspFile string, node *Node, sat chan<- bool, quit <-chan bool) {
	cmd := exec.Command(nacre, cspFile, "-complete", "-sols", "-verb=3")
	stdout, err := cmd.StdoutPipe()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err != nil {
		panic(err)
	}
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	res := false
	tuples := fetchTuples(bufio.NewReader(stdout), cspFile, node, quit)
	for tup := range tuples {
		select {
		case <-quit:
			err = cmd.Process.Kill()
			if err != nil {
				panic(err)
			}
			return
		default:
			res = true
			if _, added := node.Tuples.AddTuple(tup); !added {
				panic(fmt.Sprintf("node %v, %s: Tuple arity does not match with relation arity %v", node.ID, cspFile, len(node.Tuples.Attributes())))
			}
		}
	}
	if err := cmd.Wait(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() != 40 {
			panic(fmt.Sprintf("nacre failed on %s: %v: %s", cspFile, err, stderr.String()))
		}
	}
	sat <- res
}

func fetchTuples(r *bufio.Reader, cspFile string, node *Node, quit <-chan bool) <-chan []int {
	out := make(chan []int) // TODO buffer maybe?
	go func() {
		defer close(out)
		for {
			select {
			case <-quit:
				return
			default:
				line, eof := files.ReadLine(r)
				if eof {
					return
				}
				if strings.HasPrefix(line, "v") {
					tup := makeTuple(line, cspFile, node.bagSet)
					out <- tup
				}
			}
		}
	}()
	return out
}

func filterCtrsVars(n *Node, ctrs map[string]csp.Constraint, doms map[string]string) ([]csp.Constraint, map[string]string) {
	outCtrs := make([]csp.Constraint, 0, len(n.Cover()))
	outVars := make(map[string]string)
	for _, e := range n.Cover() {
		c := ctrs[e]
		outCtrs = append(outCtrs, c)
		for _, v := range c.Variables() {
			if _, ok := outVars[v]; !ok {
				outVars[v] = doms[v]
			}
		}
	}
	return outCtrs, outVars
}
