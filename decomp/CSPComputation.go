package decomp

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/dmlongo/callidus/ctr"
)

// SolveSubCspSeq solve the CSPs associated to a hypertree sequentially
func SolveSubCspSeq(nodes []*Node, domains map[string]string, constraints map[string]ctr.Constraint, baseDir string) bool {
	subCspFolder := makeDir(baseDir + "subs/")

	sat := true
	for _, node := range nodes {
		nodeCtrs, nodeVars := filterCtrsVars(node, constraints, domains)
		subFile := subCspFolder + "sub" + strconv.Itoa(node.ID) + ".xml"
		ctr.CreateXCSPInstance(nodeCtrs, nodeVars, subFile)
		sat = solveCSPSeq(subFile, node)
		if !sat {
			break
		}
	}
	return sat
}

func solveCSPSeq(cspFile string, node *Node) bool {
	cmd := exec.Command("./libs/nacre", cspFile, "-complete", "-sols", "-verb=3")
	out, err := cmd.StdoutPipe() // TODO why StdoutPipe() and not just Run?
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(out)
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	return readTuples(reader, node)
}

func readTuples(reader *bufio.Reader, node *Node) bool {
	solFound := false
	node.Tuples = make(Relation, 0)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF && len(line) == 0 {
			break
		}
		if strings.HasPrefix(line, "v") {
			reg := regexp.MustCompile(".*<values>(.*) </values>.*")
			tup := make([]int, len(node.Bag))
			for i, value := range strings.Split(reg.FindStringSubmatch(line)[1], " ") {
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

	sat := true
	subCspWg := &sync.WaitGroup{}
	subCspWg.Add(len(nodes))
	checkSatWg := &sync.WaitGroup{}
	checkSatWg.Add(1)
	satChan := make(chan bool, len(nodes))
	for _, node := range nodes {
		go func(node *Node) {
			nodeCtrs, nodeVars := filterCtrsVars(node, constraints, domains)
			subFile := subCspFolder + "sub" + strconv.Itoa(node.ID) + ".xml"
			ctr.CreateXCSPInstance(nodeCtrs, nodeVars, subFile)
			satChan <- solveCSPPar(subFile, node, subCspWg, satChan)
		}(node)
	}
	go func() {
		defer checkSatWg.Done()
		for sat = range satChan {
			if !sat {
				break
			}
		}
	}()
	subCspWg.Wait()
	close(satChan)
	checkSatWg.Wait()
	return sat
}

func solveCSPPar(cspFile string, node *Node, wg *sync.WaitGroup, satChan chan bool) bool {
	defer wg.Done()
	cmd := exec.Command("./libs/nacre", cspFile, "-complete", "-sols", "-verb=3")
	out, err := cmd.StdoutPipe() // TODO why StdoutPipe() and not just Run?
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(out)
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	for {
		select {
		case sat := <-satChan:
			if !sat {
				err = cmd.Process.Kill()
				if err != nil {
					panic(err)
				}
				return false
			}
		default:
			return readTuples(reader, node)
		}
	}
}

func filterCtrsVars(n *Node, ctrs map[string]ctr.Constraint, doms map[string]string) ([]ctr.Constraint, map[string]string) {
	outCtrs := make([]ctr.Constraint, 0, len(n.Cover))
	outVars := make(map[string]string)
	for _, e := range n.Cover {
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

/*
func getPossibleValues(constraint *Constraint) string {
	possibleValues := "<supports> "
	if !constraint.CType {
		possibleValues = "<conflicts> "
	}
	for _, tup := range constraint.Relation {
		possibleValues += "("
		for i := range tup {
			value := strconv.Itoa(tup[i])
			if i == len(tup)-1 {
				possibleValues += value
			} else {
				possibleValues += value + ","
			}
		}
		possibleValues += ")"
	}
	if constraint.CType {
		possibleValues += " </supports>\n"
	} else {
		possibleValues += " </conflicts>\n"
	}
	return possibleValues
}

/*func doNacreMakeFile(){
	cmd := exec.Command("make")
	cmd.Dir = "/mnt/c/Users/simon/Desktop/Università/Tesi/Programmi/CSP_Project/libs/nacre_master/core"
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}*/

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
