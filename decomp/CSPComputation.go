package decomp

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/dmlongo/callidus/ctr"
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

var valuesRegex = regexp.MustCompile(`.*<values>(.*)</values>.*`)

// SolveSubCspSeq solve the CSPs associated to a hypertree sequentially
func SolveSubCspSeq(nodes []*Node, domains map[string]string, constraints map[string]ctr.Constraint, baseDir string) bool {
	subCspFolder := makeDir(baseDir + "subs/")

	sat := true
	for _, node := range nodes {
		nodeCtrs, nodeVars := filterCtrsVars(node, constraints, domains)
		subFile := subCspFolder + "sub" + strconv.Itoa(node.ID) + ".xml"
		ctr.CreateXCSPInstance(nodeCtrs, nodeVars, subFile)
		sat = solveCSPSeq(subFile, len(nodeVars), node)
		if !sat {
			break
		}
	}
	return sat
}

func solveCSPSeq(cspFile string, numVars int, node *Node) bool {
	cmd := exec.Command(nacre, cspFile, "-complete", "-sols", "-verb=3")
	out, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(out)
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	return readTuples(reader, cspFile, numVars, node)
}

func readTuples(reader *bufio.Reader, cspFile string, arity int, node *Node) bool {
	solFound := false
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF && len(line) == 0 {
			break
		}
		if strings.HasPrefix(line, "v") {
			matches := valuesRegex.FindStringSubmatch(line)
			if len(matches) < 2 {
				panic(cspFile + ", bad values= " + line)
			}
			tup := make([]int, arity)
			for i, value := range strings.Split(strings.TrimSpace(matches[1]), " ") {
				v, err := strconv.Atoi(value)
				if err != nil {
					panic(err)
				}
				tup[i] = v
			}
			node.Tuples = append(node.Tuples, tup)
			solFound = true
		}
	}
	return solFound
}

// SolveSubCspPar solve the CSPs associated to a hypertree in parallel
func SolveSubCspPar(nodes []*Node, domains map[string]string, constraints map[string]ctr.Constraint, baseDir string) bool {
	subCspFolder := makeDir(baseDir + "subs/")

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
				ctr.CreateXCSPInstance(nodeCtrs, nodeVars, subFile)
				solveCSPPar(subFile, len(nodeVars), n, sat, quit)
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

func solveCSPPar(cspFile string, numVars int, node *Node, sat chan<- bool, quit <-chan bool) {
	cmd := exec.Command(nacre, cspFile, "-complete", "-sols", "-verb=3")
	out, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(out)
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	res := false
	tuples := fetchTuples(reader, cspFile, numVars, quit)
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
			node.Tuples = append(node.Tuples, tup)
		}
	}
	sat <- res
}

func fetchTuples(reader *bufio.Reader, cspFile string, arity int, quit <-chan bool) <-chan []int {
	out := make(chan []int) // TODO buffer maybe?
	go func() {
		defer close(out)
		for {
			select {
			case <-quit:
				return
			default:
				line, err := reader.ReadString('\n')
				if err == io.EOF && len(line) == 0 {
					return
				}
				if strings.HasPrefix(line, "v") {
					matches := valuesRegex.FindStringSubmatch(line)
					if len(matches) < 2 {
						panic(cspFile + ", bad values= " + line)
					}
					tup := make([]int, arity)
					for i, value := range strings.Split(strings.TrimSpace(matches[1]), " ") {
						v, err := strconv.Atoi(value)
						if err != nil {
							panic(err)
						}
						tup[i] = v
					}
					out <- tup
				}
			}
		}
	}()
	return out
}

func filterCtrsVars(n *Node, ctrs map[string]ctr.Constraint, doms map[string]string) ([]ctr.Constraint, map[string]string) {
	outCtrs := make([]ctr.Constraint, 0, len(n.Cover()))
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

func makeDir(name string) string {
	err := os.RemoveAll(name)
	if err != nil {
		panic(err)
	}
	err = os.Mkdir(name, 0777)
	if err != nil {
		panic(err)
	}
	return name
}

// PrintMemUsage prints the memory usage
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
