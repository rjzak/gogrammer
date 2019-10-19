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
	fmt.Fprintf(os.Stderr, "\tAvailable modes: NGRAM, INFO\n")
	for _, fset := range flags {
		fset.Usage()
	}
	fmt.Println("Usage of INFO: INFO <FilePath>")
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

	if len(os.Args) < 3 {
		PrintUsage([]flag.FlagSet{*ngrammingFlags})
	}

	start := time.Now()
	switch os.Args[1] {
		case "NGRAM":
			ngrammingFlags.Parse(os.Args[2:])
			CreateKeeplist(ngrammingFlags.Args(), *N, *toKeep, *outputFile, *threads, *useHash, *skipGram, *name)
		case "INFO":
			ShowKeeplistInfo(os.Args[2])
		default:
			PrintUsage([]flag.FlagSet{*ngrammingFlags})
	}
	duration := time.Since(start)
	fmt.Printf("Elapsed time: %s\n", duration)
}