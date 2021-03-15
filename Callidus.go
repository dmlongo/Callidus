package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/dmlongo/callidus/ctr"
	"github.com/dmlongo/callidus/decomp"
	"github.com/dmlongo/callidus/ext"
)

var csp, ht, out string
var htDebug, subInMem, subSeq, subDebug, ySeq, solDebug, printSol bool
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
	fmt.Println("done in", time.Since(startConversion))

	var rawHypertree string
	if ht == "" {
		hg := baseDir + cspName + ".hg"
		fmt.Print("Decomposing hypergraph... ")
		startDecomposition := time.Now()
		if htDebug {
			ht = baseDir + cspName + ".ht"
			rawHypertree = ext.DecomposeToFile(hg, ht)
		} else {
			rawHypertree = ext.Decompose(hg)
		}
		fmt.Println("done in", time.Since(startDecomposition))
	}

	fmt.Print("Parsing hypertree, domains and constraints... ")
	startPrep := time.Now()

	var tree decomp.Hypertree
	var root *decomp.Node
	if ht != "" || htDebug {
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

	fmt.Println("done in", time.Since(startPrep))

	//go func() {
	//	for {
	//		PrintMemUsage()
	//		time.Sleep(5 * time.Second)
	//	}
	//}()
	var satisfiable bool
	fmt.Print("Solving sub-CSPs... ")
	startSubComputation := time.Now()
	if subSeq {
		satisfiable = decomp.SolveSubCspSeq(tree, domains, constraints, baseDir)
	} else {
		satisfiable = decomp.SolveSubCspPar(tree, domains, constraints, baseDir) // TODO a bit buggy
	}
	fmt.Println("done in", time.Since(startSubComputation))
	if !satisfiable {
		fmt.Println("NO SOLUTIONS")
		return
	}
	if subDebug {
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
	fmt.Println("done in", time.Since(startYannakakis))
	fmt.Println("Callidus solved", csp, "in", time.Since(start))

	fmt.Print("Computing all solutions... ")
	startComputeAll := time.Now()
	allSolutions := decomp.ComputeAllSolutions(root)
	fmt.Println("done in", time.Since(startComputeAll))

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

	if !subDebug {
		err := os.RemoveAll(baseDir)
		if err != nil {
			panic(err)
		}
	}
}

func setFlags() {
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	flagSet.SetOutput(ioutil.Discard) //todo: see what happens without this line

	flagSet.StringVar(&csp, "csp", "", "Path to the CSP to solve (XCSP3 format)")
	flagSet.StringVar(&ht, "ht", "", "Path to a decomposition of the CSP to solve (GML format)")
	flagSet.StringVar(&out, "out", "", "Save the solutions of the CSP into the specified file")
	flagSet.BoolVar(&htDebug, "htDebug", false, "Write hypertree on disk for debug (false if -ht is set)")
	flagSet.BoolVar(&subDebug, "subDebug", false, "Write sub-CSP files on disk for debug")
	flagSet.BoolVar(&subInMem, "subInMem", false, "Activate in-memory computation of sub-CSPs")
	flagSet.BoolVar(&subSeq, "subSeq", false, "Activate sequential computation of sub-CSPs")
	flagSet.BoolVar(&ySeq, "ySeq", false, "Use sequential Yannakakis' algorithm")
	flagSet.BoolVar(&solDebug, "solDebug", false, "Check solutions of the CSP")
	flagSet.BoolVar(&printSol, "printSol", false, "Print solutions of the CSP")

	parseError := flagSet.Parse(os.Args[1:])
	if parseError != nil {
		fmt.Print("Parse Error:\n", parseError.Error(), "\n\n")
	}

	if parseError != nil || csp == "" {
		out := fmt.Sprint("Usage of Callidus (https://github.com/dmlongo/Callidus)\n")
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
	re = regexp.MustCompile("\\..*")
	cspDir = re.ReplaceAllString(cspName, "")
	baseDir = wrkdir + "/" + cspDir + "/"
	//fmt.Println("cspDir=", cspDir, ", cspName=", cspName, ", baseDir=", baseDir)

	//if ht == "" {ht = "output" + folder + "hypertree"}
	//fmt.Printf("csp=%v\nht=%v\nout=%v\n", csp, ht, out)

	contSols = 0
}
