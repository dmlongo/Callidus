package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
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
var htDebug, tabDebug, subDebug, solDebug bool
var subInMem, printSol, printTimes bool
var contSols int

const wrkdir = "wrkdir"

var cspDir string
var cspName string
var baseDir string

func main() {
	setFlags()

	fmt.Println("Callidus starts!")
	start := time.Now()

	fmt.Print("Creating hypergraph... ")
	startConversion := time.Now()
	hypergraph := ext.Convert(csp, baseDir)
	durConversion := time.Since(startConversion)
	fmt.Println("done in", durConversion)

	var rawHypertree string
	var startDecomposition time.Time
	var durDecomp time.Duration
	if ht == "" {
		hg := baseDir + cspName + ".hg"
		fmt.Print("Decomposing hypergraph... ")
		startDecomposition = time.Now()
		if htDebug {
			ht = baseDir + cspName + ".ht"
			rawHypertree = ext.DecomposeToFile(hg, ht, decompTime)
		} else {
			rawHypertree = ext.Decompose(hg, decompTime)
		}
		durDecomp = time.Since(startDecomposition)
		fmt.Println("done in", durDecomp)
	}

	fmt.Print("Parsing hypertree, domains and constraints... ")
	startParsing := time.Now()

	var tree decomp.Hypertree
	var root *decomp.Node
	if ht != "" {
		root, tree = ext.ParseDecompFromFile(ht)
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

	//go func() {
	//	for {
	//		PrintMemUsage()
	//		time.Sleep(5 * time.Second)
	//	}
	//}()
	var satisfiable bool
	fmt.Print("Solving sub-CSPs... ")
	startSubComp := time.Now()
	if subSeq {
		satisfiable = decomp.SolveSubCspSeq(tree, domains, constraints, baseDir)
	} else {
		satisfiable = decomp.SolveSubCspPar(tree, domains, constraints, baseDir) // TODO a bit buggy
	}
	durSubComp := time.Since(startSubComp)
	fmt.Println("done in", durSubComp)
	if !satisfiable {
		fmt.Println("NO SOLUTIONS")
		if printTimes {
			durs := []time.Duration{durConversion, durDecomp, durParsing, durSubComp, 0, 0, time.Since(start)}
			printDurations(durs)
		}
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

	fmt.Print("Running Yannakakis... ")
	startYannakakis := time.Now()
	if ySeq {
		decomp.YannakakisSeq(root)
	} else {
		decomp.YannakakisPar(root)
	}
	durYannakakis := time.Since(startYannakakis)
	fmt.Println("done in", durYannakakis)
	durSolving := time.Since(start)
	fmt.Println("Callidus solved", csp, "in", durSolving)

	fmt.Print("Computing all solutions... ")
	startComputeAll := time.Now()
	if ySeq {
		decomp.FullyReduceRelationsSeq(root)
	} else {
		decomp.FullyReduceRelationsPar(root)
	}
	allSolutions := decomp.ComputeAllSolutions(root) // TODO make a parallel version of this
	durComputeAll := time.Since(startComputeAll)
	durSolvingAll := time.Since(start)
	fmt.Println("done in", durComputeAll)
	contSols = len(root.Tuples)

	if solDebug {
		for _, sol := range allSolutions {
			if err, ok := ext.CheckSolution(csp, sol); !ok {
				panic(err)
				//panic(fmt.Sprintf("%v is not a solution.", sol))
			}
		}
	}

	if printSol {
		startPrinting := time.Now()
		if len(allSolutions) > 0 {
			for _, sol := range allSolutions {
				sol.Print()
			}
		} else {
			fmt.Println(csp, " has no solutions")
		}
		fmt.Println("Printing done in", time.Since(startPrinting))
	}

	if out != "" {
		// TODO write to out
	}

	if printTimes {
		durs := []time.Duration{durConversion, durDecomp, durParsing, durSubComp, durYannakakis, durComputeAll, durSolvingAll}
		printDurations(durs)
	}

	if !subDebug {
		err := os.RemoveAll(baseDir)
		if err != nil {
			panic(err)
		}
	}
}

func printDurations(durations []time.Duration) {
	var sb strings.Builder
	for _, d := range durations {
		//sb.WriteString(d.String())
		sb.WriteString(strconv.Itoa(int(d.Milliseconds())))
		sb.WriteString(";")
	}
	sb.WriteString(strconv.Itoa(contSols))
	fmt.Println(sb.String())
}

func setFlags() {
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	flagSet.SetOutput(ioutil.Discard) //todo: see what happens without this line

	flagSet.StringVar(&csp, "csp", "", "Path to the CSP to solve (XCSP3 format)")
	flagSet.StringVar(&ht, "ht", "", "Path to a decomposition of the CSP to solve (GML format)")
	flagSet.StringVar(&out, "out", "", "Save the solutions of the CSP into the specified file")
	flagSet.StringVar(&decompTime, "decompTime", "3600", "Set a timeout (seconds) for computing a decomposition of the CSP")
	flagSet.BoolVar(&htDebug, "htDebug", false, "Write hypertree on disk for debug (false if -ht is set)")
	flagSet.BoolVar(&subDebug, "subDebug", false, "Write sub-CSP files on disk for debug") // TODO update
	flagSet.BoolVar(&tabDebug, "tabDebug", false, "Save solutions of sb-CSPs on disk for debug")
	flagSet.BoolVar(&subInMem, "subInMem", false, "Activate in-memory computation of sub-CSPs")
	flagSet.BoolVar(&subSeq, "subSeq", false, "Activate sequential computation of sub-CSPs")
	flagSet.BoolVar(&ySeq, "ySeq", false, "Use sequential Yannakakis' algorithm")
	flagSet.BoolVar(&solDebug, "solDebug", false, "Check solutions of the CSP")
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

	re := regexp.MustCompile(".*/")
	cspName = re.ReplaceAllString(csp, "")
	re = regexp.MustCompile(`\..*`)
	cspDir = re.ReplaceAllString(cspName, "")
	baseDir = wrkdir + "/" + cspDir + "/"
	fmt.Println("cspDir=", cspDir, ", cspName=", cspName, ", baseDir=", baseDir)

	//if ht == "" {ht = "output" + folder + "hypertree"}
	//fmt.Printf("csp=%v\nht=%v\nout=%v\n", csp, ht, out)

	err := os.RemoveAll(baseDir) // TODO removing wastes time, not necessary
	if err != nil {
		panic(err)
	}

	contSols = 0
}
