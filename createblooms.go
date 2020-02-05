package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

func CreateBloomFilters(N, K int, fp_rate float64, dirs []string, output string) {
	bloom_slots, bloom_hashes := CalcSlotsHashes(K, fp_rate)

	filesList := make([]string, 0)
	for _, sampleDir := range dirs {
		err := filepath.Walk(sampleDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.Mode().IsRegular() {
				filesList = append(filesList, path)
			} else {
				if info.Mode()&os.ModeSymlink > 0 {
					pathLink, err := filepath.EvalSymlinks(path)
					if err != nil {
						return err
					}
					pathInfo, err := os.Lstat(pathLink)
					if err != nil {
						return err
					}
					if pathInfo.Mode().IsRegular() {
						filesList = append(filesList, path)
					}
				}
			}

			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking %s: %v.\n", sampleDir, err)
		}
	}
	numFiles := len(filesList)
	fmt.Printf("Found %d files.\n", len(filesList))
	if numFiles < 1 {
		return
	}

	var lock sync.Mutex
	var wg sync.WaitGroup
	hashNgrams := NewMaxIntArray()
	threads := int(math.Max(float64(runtime.NumCPU()-1), 4))
	numPieces := int(numFiles / threads) - 1

	fmt.Printf("Collecting %d-grams.\n", N)
	for i := 0; i < numPieces; i++ {
		start := numPieces * i
		end := start + numPieces
		if end >= numFiles {
			end = numFiles - 1
		}
		if start > numFiles {
			break
		}
		thisSlice := filesList[start:end]
		go HashNgramming(thisSlice, N, &hashNgrams, 1, &lock, &wg)
		wg.Add(1)
	}

	var indiciesToValuesArray [][]uint32
	for index, value := range hashNgrams {
		if value > uint32(numFiles/2.0) {
			indiciesToValuesArray = append(indiciesToValuesArray, []uint32{uint32(index), value})
		}
	}

	fmt.Println("Sorting ngrams.")
	QuickSelect(UIntArraySlice32(indiciesToValuesArray), K+2)
	indiciesToValuesArray = indiciesToValuesArray[:int(math.Min(float64(K+2), float64(len(indiciesToValuesArray))))]

	var topValues []uint32
	for _, val := range indiciesToValuesArray {
		topValues = append(topValues, val[0])
	}

	runtime.GC() // Maybe we can garbage collect the old hashNgrams array
	fmt.Println("Performing second run to get ngrams from hashes.")

	var keepers [][]byte
	for i := 0; i < numPieces; i++ {
		start := numPieces * i
		end := start + numPieces
		if end >= numFiles {
			end = numFiles - 1
		}
		if start > numFiles {
			break
		}
		thisSlice := filesList[start:end]
		go HashesToNgrams(thisSlice, &topValues, N, K, &lock, &keepers, &wg)
		wg.Add(1)
	}
	wg.Wait()

	bloom := NewBloom(bloom_slots, bloom_hashes)
	for _, ngram := range keepers {
		err := bloom.Put(ngram, 1)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to store byte ngram in bloom filter: %v.\n", err)
			os.Exit(1)
		}
	}

	bloom.Save(output)
}