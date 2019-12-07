package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"
)

func PrintUsage(flags []flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Usage: %s MODE ARGS\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\tAvailable modes: NGRAM, LISTINFO, LISTCOMPARE, DATASET\n")
	fmt.Fprintf(os.Stderr, "A simple tool for creating malware/goodware datasets from raw byte.\nhttps://github.com/rjzak/gogrammer/\n\n")
	for _, fset := range flags {
		fset.Usage()
	}
	fmt.Println("Usage of INFO: LISTINFO <FilePath>")
	fmt.Printf("Go: %s\n", runtime.Version())
	os.Exit(1)
}

func main() {
	var ngrammingFlags = flag.NewFlagSet("NGRAM", flag.ExitOnError)
	var N = ngrammingFlags.Int("size", 6, "Size of ngrams (value of N)")
	var toKeep = ngrammingFlags.Int("keep", 1000, "Number of top ngrams to keep")
	var useHash = ngrammingFlags.Bool("hash", false, "Use hash-grams")
	var skipGram = ngrammingFlags.Uint("skip", 1, "Skip-grams when using hashing")
	var name = ngrammingFlags.String("name", "unnamed", "Name of the data represented")
	var threads = ngrammingFlags.Int("threads", runtime.NumCPU(), "Number of threads to use")
	var outputFile = ngrammingFlags.String("output", "output.grams", "Output file name")

	var listCompareFlags = flag.NewFlagSet("LISTCOMPARE", flag.ExitOnError)
	var listOne = listCompareFlags.String("1", "list1.out", "First list for comparison")
	var listTwo = listCompareFlags.String("2", "list2.out", "Second list for comparison")

	var makeDatasetFlags = flag.NewFlagSet("DATASET", flag.ExitOnError)
	var keepListPath = makeDatasetFlags.String("kl", "keeplist.out", "Keep List file")
	var goodwarePath = makeDatasetFlags.String("goodware", "goodware", "Path to goodware directory")
	var malwarePath = makeDatasetFlags.String("malware", "malware", "Path to malware directory")
	var datasetOutputPath = makeDatasetFlags.String("dataset", "dataset.csv", "Path for resulting dataset file")
	var dsetThreads = makeDatasetFlags.Int("threads", runtime.NumCPU(), "Number of threads to use")

	flagsArray := []flag.FlagSet{*ngrammingFlags, *listCompareFlags, *makeDatasetFlags}
	if len(os.Args) < 3 {
		PrintUsage(flagsArray)
	}

	start := time.Now()
	switch os.Args[1] {
		case "NGRAM":
			ngrammingFlags.Parse(os.Args[2:])
			CreateKeeplist(ngrammingFlags.Args(), *N, *toKeep, *outputFile, *threads, *useHash, *skipGram, *name)
		case "LISTINFO":
			ShowKeeplistInfo(os.Args[2])
	    case "LISTCOMPARE":
	    	listCompareFlags.Parse(os.Args[2:])
			KeepListCompare(*listOne, *listTwo)
		case "DATASET":
			makeDatasetFlags.Parse(os.Args[2:])
			CreateDataset(*malwarePath, *goodwarePath, *keepListPath, *datasetOutputPath, *dsetThreads)
		default:
			PrintUsage(flagsArray)
	}
	duration := time.Since(start)
	fmt.Printf("Elapsed time: %s\n", duration)
}