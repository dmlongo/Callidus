package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/dmlongo/callidus/ctr"
	"github.com/dmlongo/callidus/decomp"
	"github.com/dmlongo/callidus/ext"
)

var csp, ht, out string
var decompTime string
var subSeq, ySeq bool
var htDebug, tabDebug, subDebug, memDebug, solDebug bool
var subInMem, printRel, printSol, printTimes bool
var all bool

var start time.Time
var durs []time.Duration
var solutions []ctr.Solution
var numSols int

const wrkdir = "wrkdir"

var cspDir string
var cspName string
var baseDir string

func main() {
	defer cleanup()
	setFlags()

	fmt.Printf("Callidus starts solving %s!\n", cspName)
	start = time.Now()

	fmt.Print("Creating hypergraph... ")
	startConversion := time.Now()
	hypergraph := ext.Convert(csp, baseDir)
	durConversion := time.Since(startConversion)
	fmt.Println("done in", durConversion)
	durs = append(durs, durConversion)

	var rawHypertree string
	var startDecomposition time.Time
	var durDecomp time.Duration
	if ht == "" {
		hg := baseDir + cspName + ".hg"
		fmt.Print("Decomposing hypergraph... ")
		startDecomposition = time.Now()
		if htDebug {
			//ht = baseDir + cspName + ".ht"
			rawHypertree = ext.DecomposeToFile(hg, baseDir+cspName+".ht", decompTime)
		} else {
			rawHypertree = ext.Decompose(hg, decompTime)
		}
		durDecomp = time.Since(startDecomposition)
		fmt.Println("done in", durDecomp)
	}
	durs = append(durs, durDecomp)

	if ht == "" && rawHypertree == "" {
		fmt.Printf("Could not find any decomposition in %vs\n", decompTime)
		return
	}

	fmt.Print("Parsing hypertree, domains and constraints... ")
	startParsing := time.Now()

	var tree decomp.Hypertree
	var root *decomp.Node
	if ht != "" {
		root, tree = ext.ParseDecompFromFile(ht)
		// TODO check if decomp is correct for hg
	} else {
		root, tree = ext.ParseDecomp(&rawHypertree)
	}
	tree.Complete(hypergraph)

	var domains map[string]string
	domFile := baseDir + cspName + ".dom"
	domains = ext.ParseDomains(domFile)

	var constraints map[string]ctr.Constraint
	ctrFile := baseDir + cspName + ".ctr"
	constraints = ext.ParseConstraints(ctrFile)

	durParsing := time.Since(startParsing)
	fmt.Println("done in", durParsing)
	durs = append(durs, durParsing)

	if memDebug {
		go func() {
			for {
				PrintMemUsage()
				time.Sleep(5 * time.Second)
			}
		}()
	}

	var satisfiable bool
	fmt.Print("Solving sub-CSPs... ")
	startSubComp := time.Now()
	if subSeq {
		satisfiable = decomp.SolveSubCspSeq(tree, domains, constraints, baseDir)
	} else {
		satisfiable = decomp.SolveSubCspPar(tree, domains, constraints, baseDir)
	}
	durSubComp := time.Since(startSubComp)
	fmt.Println("done in", durSubComp)
	durs = append(durs, durSubComp)
	if !satisfiable {
		printOutput(satisfiable)
		return
	}

	if tabDebug {
		tablesFolder := baseDir + "tables/"
		err := os.RemoveAll(tablesFolder)
		if err != nil {
			panic(err)
		}
		err = os.Mkdir(tablesFolder, 0777)
		if err != nil {
			panic(err)
		}
		for _, node := range tree {
			tableFile := tablesFolder + "sub" + strconv.Itoa(node.ID) + ".tab"
			ext.CreateSolutionTable(tableFile, node)
		}
	}

	if printRel {
		decomp.PrintTreeRelations(root)
	}

	fmt.Print("Running Yannakakis... ") // TODO the csp can be unsat also here
	startYannakakis := time.Now()
	if ySeq {
		decomp.YannakakisSeq(root)
	} else {
		decomp.YannakakisPar(root)
	}
	durYannakakis := time.Since(startYannakakis)
	fmt.Println("done in", durYannakakis)
	durs = append(durs, durYannakakis)
	if !satisfiable {
		printOutput(satisfiable)
		return
	}

	if printRel {
		decomp.PrintTreeRelations(root)
	}

	if all {
		fmt.Print("Computing all solutions... ")
		startComputeAll := time.Now()
		if ySeq {
			decomp.FullyReduceRelationsSeq(root)
		} else {
			decomp.FullyReduceRelationsPar(root)
		}

		if printRel {
			decomp.PrintTreeRelations(root)
		}

		if ySeq {
			solutions = decomp.ComputeAllSolutionsSeq(root)
		} else {
			solutions = decomp.ComputeAllSolutionsPar(root)
		}
		durComputeAll := time.Since(startComputeAll)
		fmt.Println("done in", durComputeAll)
		durs = append(durs, durComputeAll)

		if printRel {
			decomp.PrintTreeRelations(root)
		}
	}

	printOutput(true)

	if solDebug {
		fmt.Print("Checking ", numSols, " solutions... ")
		startCheckSol := time.Now()
		for _, sol := range solutions {
			if err, ok := ext.CheckSolution(csp, sol); !ok {
				panic(fmt.Sprintf("%v is not a solution: %v", sol, err))
			}
		}
		fmt.Println("done in", time.Since(startCheckSol))
	}

	if out != "" {
		// TODO write to out
	}
}

func printOutput(sat bool) {
	durCallidus := time.Since(start)
	numSols = len(solutions)

	if !all {
		if !sat {
			fmt.Println(csp, "has no solutions")
		} else {
			fmt.Println(csp, "has at least one solution")
		}
		fmt.Println("Callidus solved", csp, "in", durCallidus)
	} else {
		fmt.Println("Callidus found", numSols, "solutions in", durCallidus)
	}

	if printTimes {
		//durs := []time.Duration{durConversion, durDecomp, durParsing, durSubComp, durYannakakis, durComputeAll, durSolvingAll}
		for i := len(durs); i < 6; i++ {
			durs = append(durs, 0)
		}
		durs = append(durs, durCallidus)

		var sb strings.Builder
		for _, d := range durs {
			sb.WriteString(strconv.Itoa(int(d.Milliseconds())))
			sb.WriteString(";")
		}

		if !all {
			if !sat {
				sb.WriteString("n")
			} else {
				sb.WriteString("y")
			}
		} else {
			sb.WriteString(strconv.Itoa(numSols))
		}

		fmt.Println("convert;decomp;parsing;subcsp;yanna;compall;total;sols")
		fmt.Println(sb.String())
	}
}

func setFlags() {
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	flagSet.SetOutput(ioutil.Discard) //todo: see what happens without this line

	flagSet.StringVar(&csp, "csp", "", "Path to the CSP to solve (XCSP3 format)")
	flagSet.StringVar(&ht, "ht", "", "Path to a decomposition of the CSP to solve (GML format)")
	flagSet.StringVar(&out, "out", "", "Save the solutions of the CSP into the specified file")
	flagSet.StringVar(&decompTime, "decompTime", "3600", "Set a timeout (seconds) for computing a decomposition of the CSP")
	flagSet.BoolVar(&all, "all", false, "Compute all solutions of the CSP")
	flagSet.BoolVar(&htDebug, "htDebug", false, "Write hypertree on disk for debug (false if -ht is set)")
	flagSet.BoolVar(&subDebug, "subDebug", false, "Write sub-CSP files on disk for debug") // TODO update
	flagSet.BoolVar(&tabDebug, "tabDebug", false, "Save solutions of sb-CSPs on disk for debug")
	flagSet.BoolVar(&memDebug, "memDebug", false, "Print memory usage every 5sseconds")
	flagSet.BoolVar(&subInMem, "subInMem", false, "Activate in-memory computation of sub-CSPs")
	flagSet.BoolVar(&subSeq, "subSeq", false, "Activate sequential computation of sub-CSPs")
	flagSet.BoolVar(&ySeq, "ySeq", false, "Use sequential Yannakakis' algorithm")
	flagSet.BoolVar(&solDebug, "solDebug", false, "Check solutions of the CSP")
	flagSet.BoolVar(&printRel, "printRel", false, "Print relations at every step of the CSP resolution")
	flagSet.BoolVar(&printSol, "printSol", false, "Print solutions of the CSP")
	flagSet.BoolVar(&printTimes, "printTimes", false, "Print times of each resolution phase")

	parseError := flagSet.Parse(os.Args[1:])
	if parseError != nil {
		fmt.Print("Parse Error:\n", parseError.Error(), "\n\n")
	}

	if parseError != nil || csp == "" {
		out := "Usage of Callidus (https://github.com/dmlongo/Callidus)\n"
		flagSet.VisitAll(func(f *flag.Flag) {
			if f.Name != "csp" {
				return
			}
			s := fmt.Sprintf("%T", f.Value) // used to get type of flag
			if s[6:len(s)-5] != "bool" {
				out += fmt.Sprintf("  -%-10s \t<%s>\n", f.Name, s[6:len(s)-5])
			} else {
				out += fmt.Sprintf("  -%-10s \n", f.Name)
			}
			out += fmt.Sprintln("\t" + f.Usage)
		})
		out += fmt.Sprintln("\nOptional Arguments: ")
		flagSet.VisitAll(func(f *flag.Flag) {
			if f.Name == "csp" {
				return
			}
			s := fmt.Sprintf("%T", f.Value) // used to get type of flag
			if s[6:len(s)-5] != "bool" {
				out += fmt.Sprintf("  -%-10s \t<%s>\n", f.Name, s[6:len(s)-5])
			} else {
				out += fmt.Sprintf("  -%-10s \n", f.Name)
			}
			out += fmt.Sprintln("\t" + f.Usage)
		})
		fmt.Fprintln(os.Stderr, out)

		os.Exit(1)
	}

	if ht != "" {
		htDebug = false
	}

	re := regexp.MustCompile(`.*/`)
	cspName = re.ReplaceAllString(csp, "")
	re = regexp.MustCompile(`\..*`)
	cspDir = re.ReplaceAllString(cspName, "")
	baseDir = wrkdir + "/" + cspDir + "/"
	//fmt.Println("cspDir=", cspDir, ", cspName=", cspName, ", baseDir=", baseDir)

	//if ht == "" {ht = "output" + folder + "hypertree"}
	//fmt.Printf("csp=%v\nht=%v\nout=%v\n", csp, ht, out)

	err := os.RemoveAll(baseDir) // TODO removing wastes time, not necessary
	if err != nil {
		panic(err)
	}

	numSols = 0
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

func cleanup() {
	if !subDebug {
		err := os.RemoveAll(baseDir)
		if err != nil {
			panic(err)
		}
	}
}
