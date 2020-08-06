package pre_processing

import "regexp"

type Settings struct {
	ParallelYannakaki bool
	InMemory          bool
	Solver            string
	Debug             bool
	ParallelSC        bool
	BalancedAlgorithm string
	PrintSol          bool
	ComputeWidth      bool
	Output            string
	HypertreeFile     string
	FolderName        string
}

var SystemSettings *Settings = &Settings{}

func (settings *Settings) InitSettings(args []string, filePath string) {
	settings.ParallelYannakaki = selectYannakakiVersion(args) //true if parallel, false if sequential
	settings.InMemory = selectComputation(args)               // -i for computation in memory, inMemory = true or inMemory = false if not
	settings.Debug = selectDebugOption(args)
	settings.ParallelSC = selectSubComputationExec(args)
	settings.PrintSol = selectPrintSol(args)
	settings.Output = writeSolution(args)
	settings.FolderName = getFolderName(filePath)
	settings.HypertreeFile = takeHypertreeFile(args, "output"+settings.FolderName)
}

func contains(args []string, param string) int {
	for i := 0; i < len(args); i++ {
		if args[i] == param {
			return i
		}
	}
	return -1
}

func takeHypertreeFile(args []string, folderName string) string {
	i := contains(args, "-h")
	if i == -1 {
		i = contains(args, "--hypertree")
	}
	if i != -1 {
		return args[i+1]
	}
	return folderName + "hypertree"
}

func selectYannakakiVersion(args []string) bool {
	i := contains(args, "-y")
	if i == -1 {
		i = contains(args, "--yannakaki")
	}
	if i != -1 {
		version := args[i+1]
		if version != "s" && version != "p" {
			panic(args[i] + " must be followed by 's' or 'p'")
		}
		if version == "p" {
			return true
		}
	}
	return false
}

func selectComputation(args []string) bool {
	i := contains(args, "-i")
	if i != -1 {
		return true
	}
	return false
}

func selectDebugOption(args []string) bool {
	if contains(args, "-d") != -1 || contains(args, "--debug") != -1 {
		return true
	}
	return false
}

func selectSubComputationExec(args []string) bool {
	if i := contains(args, "-sc"); i != -1 {
		if args[i+1] != "p" && args[i+1] != "s" {
			panic(args[i] + " must be followed by 's' or 'p'")
		}
		if args[i+1] == "s" {
			return false
		}
	}
	return true
}

func selectPrintSol(args []string) bool {
	if i := contains(args, "-printSol"); i != -1 {
		if args[i+1] != "yes" && args[i+1] != "no" {
			panic(args[i] + " must be followed by 'yes' or 'no'")
		}
		if args[i+1] == "no" {
			return false
		}
	}
	return true
}

func writeSolution(args []string) string {
	if i := contains(args, "-output"); i != -1 {
		return args[i+1]
	}
	return ""
}

func getFolderName(filePath string) string {
	re := regexp.MustCompile(".*/")
	folderName := re.ReplaceAllString(filePath, "")
	re = regexp.MustCompile("\\..*")
	folderName = re.ReplaceAllString(folderName, "")
	folderName = folderName + "/"
	return folderName
}
