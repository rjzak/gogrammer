package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

const MaxUint = ^uint32(0) -1

// defining the array with MaxUint size causes compiler error
type MaxIntArray []uint32

func NewMaxIntArray() MaxIntArray {
	var a MaxIntArray
	/*var i uint32
	for i = 0; i < MaxUint; i++ {
		a = append(a, 0)
	}*/
	a = make([]uint32, 500000000)
	return a
}

func (a MaxIntArray) Len() int {
	return len(a)
}

func (a MaxIntArray) Less (i, j int) bool {
	return a[i] < a[j]
}

func (a MaxIntArray) Swap(i, j int) {
	temp := a[i]
	a[i] = a[j]
	a[j] = temp
}

func (a MaxIntArray) Top(x int) []uint32 {
	temp := []uint32{}
	for i := 0; i < x; i++ {
		temp = append(temp, a[i])
	}
	return temp
}

func BasicNgramming(fileList []string, ngramSize int, ngrams *AtomicByteMap, lock *sync.Mutex, wg *sync.WaitGroup) {
	for _, filePath := range fileList {
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			lock.Lock()
			fmt.Printf("Unable to open %s.\n", filePath)
			lock.Unlock()
			continue
		}
		current := 0
		for {
			gram := content[current: current+ngramSize]
			//fmt.Printf("%s: %s\n", filePath, gram)
			ngrams.IncrementByte(gram)
			current += 1
			if current + ngramSize >= len(content) {
				break
			}
		}
	}
	wg.Done()
}

func HashNgramming(fileList []string, ngramSize int, hashArray *MaxIntArray, skipGram uint32, lock *sync.Mutex, wg *sync.WaitGroup) {
	for _, filePath := range fileList {
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			lock.Lock()
			fmt.Printf("Unable to open %s.\n", filePath)
			lock.Unlock()
			continue
		}
		current := 0
		for {
			CastagnoliTable := crc32.MakeTable(crc32.Castagnoli)
			gram := content[current: current+ngramSize]

			if binary.BigEndian.Uint32(gram) % skipGram == 0 {
				gramIndex := crc32.Checksum(gram, CastagnoliTable)/100.0
				lock.Lock()
				(*hashArray)[gramIndex] += 1
				lock.Unlock()
			}

			current += 1
			if current + ngramSize >= len(content) {
				break
			}
		}
	}
	wg.Done()
}

func HashesToNgrams(fileList []string, hashArray *MaxIntArray, ngramSize int, lock *sync.Mutex, ngrams *AtomicByteMap, wg *sync.WaitGroup) {
	for _, filePath := range fileList {
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			lock.Lock()
			fmt.Printf("Unable to open %s.\n", filePath)
			lock.Unlock()
			continue
		}
		current := 0
		for {
			CastagnoliTable := crc32.MakeTable(crc32.Castagnoli)
			gram := content[current: current+ngramSize]
			gramIndex := crc32.Checksum(gram, CastagnoliTable)/100.0
			if (*hashArray)[gramIndex] > 1 {
				ngrams.IncrementByte(gram)
			}

			current += 1
			if current + ngramSize >= len(content) {
				break
			}
		}
	}
	wg.Done()
}

func CreateKeeplist(filePaths []string, ngramSize int, numGramsToKeep int, outputFile string, threads int, useHash bool, skipgram uint, name string) {
	runtime.GOMAXPROCS(threads)

	if ngramSize < 2 {
		fmt.Fprintf(os.Stderr, "Bytes less than two does not make sense.\n")
		os.Exit(1)
	}

	if numGramsToKeep < 2 {
		fmt.Fprintf(os.Stderr, "To-keep is too small.\n")
		os.Exit(1)
	}

	if exists(outputFile) {
		fmt.Fprintf(os.Stderr, "Output file already exists, won't overwrite.\n")
		os.Exit(1)
	}

	var fileList = []string{}

	fmt.Println("Collecting files.")
	for _, input := range filePaths {
		if info, err := os.Stat(input); err == nil && info.IsDir() {
			err = filepath.Walk(input, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.Mode().IsRegular() {
					fileList = append(fileList, path)
				} else {
					path, err := filepath.EvalSymlinks(path)
					if err == nil {
						info, err = os.Stat(path)
						if err == nil && info.Mode().IsRegular() {
							fileList = append(fileList, path)
						}
					}
				}
				return nil
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "error walking the path %q: %v\n", input, err)
				return
			}
		} else {
			if fi, err := os.Stat(input); err == nil && fi.Mode().IsRegular() {
				fileList = append(fileList, input)
			}
		}
	}

	numFiles := len(fileList)
	if numFiles == 0 {
		fmt.Println("No files found.")
		return
	}
	if useHash {
		fmt.Printf("Starting hashing of %d files.\n", numFiles)
	} else {
		fmt.Printf("Starting ngramming %d files.\n", numFiles)
	}

	var regularNgrams = NewAtomicByteMap()
	var hashNgrams MaxIntArray
	if useHash {
		hashNgrams = NewMaxIntArray()
	}

	var lock sync.Mutex
	var numPieces = int(numFiles / threads) - 1
	var wg sync.WaitGroup
	if numPieces < 2 {
		// Not enough files for threading
		wg.Add(1)
		if useHash {
			HashNgramming(fileList, ngramSize, &hashNgrams, uint32(skipgram), &lock, &wg)
		} else {
			BasicNgramming(fileList, ngramSize, &regularNgrams, &lock, &wg)
		}
		wg.Wait()
	} else {
		for i := 0; i < numPieces; i++ {
			start := numPieces * i
			end := start + numPieces
			if end >= numFiles {
				end = numFiles - 1
			}
			if start > numFiles {
				break
			}
			thisSlice := fileList[start:end]
			wg.Add(1)
			if useHash {
				go HashNgramming(thisSlice, ngramSize, &hashNgrams, uint32(skipgram), &lock, &wg)
			} else {
				go BasicNgramming(thisSlice, ngramSize, &regularNgrams, &lock, &wg)
			}
		}
		wg.Wait()
	}

	runtime.GC()
	fmt.Println("Sorting ngrams.")
	var keepers [][]byte
	if useHash {
		fmt.Println("Performing second run to get ngrams from hashes.")
		if numPieces < 2 {
			// Not enough files for threading
			wg.Add(1)
			HashesToNgrams(fileList, &hashNgrams, ngramSize, &lock, &regularNgrams, &wg)
		} else {
			for i := 0; i < numPieces; i++ {
				start := numPieces * i
				end := start + numPieces
				if end >= numFiles {
					end = numFiles - 1
				}
				if start > numFiles {
					break
				}
				thisSlice := fileList[start:end]
				wg.Add(1)
				go HashesToNgrams(thisSlice, &hashNgrams, ngramSize, &lock, &regularNgrams, &wg)
			}
		}
		wg.Wait()
	}

	if regularNgrams.IsEmpty() {
		fmt.Fprintf(os.Stderr, "Ngrams container empty, this won't end well.\n")
	}
	sortedGrams := regularNgrams.SortedValues()
	prev := sortedGrams[0].Value
	sortIsBroken := false
	for _, pair := range sortedGrams {
		keepers = append(keepers, pair.Key)
		if pair.Value > prev {
			sortIsBroken = true
		}
		if len(keepers) == numGramsToKeep {
			break
		}
	}
	regularNgrams.Erase()
	runtime.GC()

	if sortIsBroken {
		fmt.Println("Sorting of ngram counts not working.");
	}

	if len(keepers) != numGramsToKeep {
		fmt.Fprintf(os.Stderr, "Only found %d unique ngrams despite the request for %d ngrams.\n", len(keepers), numGramsToKeep)
	} else {
		fmt.Printf("Saving the top %d %d-grams to %s.\n", numGramsToKeep, ngramSize, outputFile)
	}

	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open %s for writing: %v.\n", outputFile, err)
		os.Exit(1)
	}
	defer file.Close()

	err = binary.Write(file, binary.BigEndian, MAGIC)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing magic number to %s: %v\n", outputFile, err)
	}

	bytesTemp := make([]byte, 4)
	binary.BigEndian.PutUint32(bytesTemp, uint32(len(name)))
	err = binary.Write(file, binary.BigEndian, bytesTemp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing name length to %s: %v\n", outputFile, err)
	}

	err = binary.Write(file, binary.BigEndian, []byte(name))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing dataset name to %s: %v\n", outputFile, err)
	}

	binary.BigEndian.PutUint32(bytesTemp, uint32(ngramSize))
	err = binary.Write(file, binary.BigEndian, bytesTemp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing ngram size to %s: %v\n", outputFile, err)
	}

	binary.BigEndian.PutUint32(bytesTemp, uint32(skipgram))
	err = binary.Write(file, binary.BigEndian, bytesTemp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing skip gram size to %s: %v\n", outputFile, err)
	}

	binary.BigEndian.PutUint32(bytesTemp, uint32(len(keepers)))
	err = binary.Write(file, binary.BigEndian, bytesTemp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing number of ngrams to %s: %v\n", outputFile, err)
	}

	justOnce := false
	for _, gram := range keepers {
		err = binary.Write(file, binary.BigEndian, gram)
		if err != nil && !justOnce {
			fmt.Fprintf(os.Stderr, "Error writing gram value to %s: %v\n", outputFile, err)
			justOnce = true
		}
	}
}