package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

func CreateLibSVMDataset(fileList []string, isMalware bool, kl *KeepList, outFile *os.File, lock *sync.Mutex, wg *sync.WaitGroup) {
	for _, filePath := range fileList {
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			lock.Lock()
			fmt.Printf("Unable to open %s.\n", filePath)
			lock.Unlock()
			continue
		}
		current := uint32(0)
		outputArray := make([]int, kl.NumNgrams)
		for {
			gram := content[current: current+kl.NgramSize]
			ngramIndex := kl.ContainsGram(gram)
			if ngramIndex > 0 {
				outputArray[ngramIndex] = 1
			}
			current += 1
			if current + kl.NgramSize >= uint32(len(content)) {
				break
			}
		}
		outputString := ""
		if isMalware {
			outputString = outputString + "1:"
		} else {
			outputString = outputString + "-1:"
		}

		for idx, val := range outputArray {
			if val > 0 {
				outputString = outputString + fmt.Sprintf(" %d:1", idx)
			}
		}

		outputString = outputString + "\n"

		lock.Lock()
		outFile.WriteString(outputString)
		lock.Unlock()
	}
	wg.Done()
}

func CreateCSVDataset(fileList []string, isMalware bool, kl *KeepList, outFile *os.File, lock *sync.Mutex, wg *sync.WaitGroup) {
	for _, filePath := range fileList {
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			lock.Lock()
			fmt.Printf("Unable to open %s.\n", filePath)
			lock.Unlock()
			continue
		}
		current := uint32(0)
		outputArray := make([]int, kl.NumNgrams)
		for {
			gram := content[current: current+kl.NgramSize]
			ngramIndex := kl.ContainsGram(gram)
			if ngramIndex > 0 {
				outputArray[ngramIndex] = 1
			}
			current += 1
			if current + kl.NgramSize >= uint32(len(content)) {
				break
			}
		}
		outputString := ""
		for _, val := range outputArray {
			if val > 0 {
				outputString = outputString + "1,"
			} else {
				outputString = outputString + "0,"
			}
		}
		if isMalware {
			outputString = outputString + "1\n"
		} else {
			outputString = outputString + "-1\n"
		}
		lock.Lock()
		outFile.WriteString(outputString)
		lock.Unlock()
	}
	wg.Done()
}

func CreateDataset(malwarePath string, goodwarePath string, keeplistPath string, outputPath string, numThreads int) {
	runtime.GOMAXPROCS(numThreads)
	var err error
	var haveErrors bool = false

	if exists(outputPath) {
		fmt.Fprintf(os.Stderr, "Output file already exists, won't overwrite.\n")
		haveErrors = true
	}

	dirInfo, err := os.Stat(malwarePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Malware directory %s is not a valid path.\n", malwarePath)
		haveErrors = true
	} else {
		if !dirInfo.IsDir() {
			malwarePath, _ = filepath.EvalSymlinks(malwarePath)
			dirInfo, err = os.Lstat(malwarePath)
			if !dirInfo.IsDir() {
				fmt.Fprintf(os.Stderr, "Malware directory %s is not a directory.\n", malwarePath)
				haveErrors = true
			}
		}
	}

	dirInfo, err = os.Stat(goodwarePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Goodware directory %s is not a valid path.\n", goodwarePath)
		haveErrors = true
	} else {
		if !dirInfo.IsDir() {
			malwarePath, _ = filepath.EvalSymlinks(goodwarePath)
			dirInfo, err = os.Lstat(goodwarePath)
			if !dirInfo.IsDir() {
				fmt.Fprintf(os.Stderr, "Goodware directory %s is not a directory.\n", goodwarePath)
				haveErrors = true
			}
		}
	}

	kl, err := ParseKeepList(keeplistPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read keeplist %s: %v.\n", keeplistPath, err)
		haveErrors = true
	}

	if haveErrors {
		return
	}

	fmt.Println("Collecting files.")

	var malwareList []string
	err = filepath.Walk(malwarePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			malwareList = append(malwareList, path)
		} else {
			path, err := filepath.EvalSymlinks(path)
			if err == nil {
				info, err = os.Stat(path)
				if err == nil && info.Mode().IsRegular() {
					malwareList = append(malwareList, path)
				}
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking malware path %s: %v\n", malwareList, err)
		return
	}

	var goodwareList []string
	err = filepath.Walk(goodwarePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			goodwareList = append(goodwareList, path)
		} else {
			path, err := filepath.EvalSymlinks(path)
			if err == nil {
				info, err = os.Stat(path)
				if err == nil && info.Mode().IsRegular() {
					goodwareList = append(goodwareList, path)
				}
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking goodware path %s: %v\n", goodwarePath, err)
		return
	}

	fmt.Printf("Found %d malware files, %d goodware files; keeplist contains %dx %d-grams.\n", len(malwareList), len(goodwareList), kl.NumNgrams, kl.NgramSize)

	file, err := os.Create(outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open %s for writing: %v.\n", outputPath, err)
		return
	}
	defer file.Close()

	var numFiles = len(malwareList)
	var numPieces = int(numFiles / numThreads) - 1
	var wg sync.WaitGroup
	var lock sync.Mutex

	datasetFunc := CreateCSVDataset
	if outputPath[len(outputPath)-7:] == ".libsvm" {
		fmt.Println("Generating libsvm dataset.")
		datasetFunc = CreateLibSVMDataset
	} else {
		fmt.Println("Generating CSV dataset.")
		for _, gram := range kl.NGrams {
			dst := make([]byte, hex.EncodedLen(len(gram)))
			hex.Encode(dst, gram)
			file.WriteString(string(dst)+",")
		}
		file.WriteString("Label\n")
	}

	if numPieces < 2 {
		// Not enough files for threading
		wg.Add(1)
		datasetFunc(malwareList, true, &kl, file, &lock, &wg)
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
			thisSlice := malwareList[start:end]
			wg.Add(1)
			go datasetFunc(thisSlice, true, &kl, file, &lock, &wg)
		}
		wg.Wait()
	}

	numFiles = len(goodwareList)
	numPieces = int(numFiles / numThreads) - 1

	if numPieces < 2 {
		// Not enough files for threading
		wg.Add(1)
		datasetFunc(goodwareList, false, &kl, file, &lock, &wg)
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
			thisSlice := goodwareList[start:end]
			wg.Add(1)
			go datasetFunc(thisSlice, false, &kl, file, &lock, &wg)
		}
		wg.Wait()
	}
}
