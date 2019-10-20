package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
)

const (
	NormalNgramming = 0x00
	HashGramming = 0x01
	UnknownNgramMethod = 0xFF
)

type KeepList struct {
	DatasetName string
	NgramSize uint32
	SkipSize uint32
	NumNgrams uint32
	CollectionMethod byte
	NGrams [][]byte
}

func (kl KeepList) Print() {
	fmt.Printf("Dataset Name: %s\n", kl.DatasetName)
	fmt.Printf("Ngram Size: %d\n", kl.NgramSize)
	fmt.Printf("Skipgram: %d\n", kl.SkipSize)
	fmt.Printf("Number of Ngrams: %d\n", kl.NumNgrams)

	switch kl.CollectionMethod {
	case NormalNgramming:
		fmt.Println("Normal ngram collection method used.")
	case HashGramming:
		fmt.Println("Hash-gramming collection used.")
	default:
		fmt.Println("Unknown collection method used.")
	}
}

func (kl KeepList) ContainsGram(g []byte) int {
	for idx, gram := range kl.NGrams {
		if bytes.Compare(gram, g) == 0 {
			return idx
		}
	}
	return -1
}

func ParseKeepList(filePath string) (KeepList, error) {
	var kl KeepList

	file, err := os.Open(filePath)
	if err != nil {
		return kl, err
	}
	defer file.Close()

	bytesTemp := make([]byte, 8)
	err = binary.Read(file, binary.BigEndian, bytesTemp)
	noMatch := false
	for i := 0; i < 8; i++ {
		if bytesTemp[i] != MAGIC[i] {
			noMatch = true
			break
		}
	}
	if noMatch {
		return kl, errors.New(fmt.Sprintf( "%s is not a keeplist.\n", filePath))
	}

	bytesTemp = make([]byte, 4)
	err = binary.Read(file, binary.BigEndian, bytesTemp)
	nameLength := binary.BigEndian.Uint32(bytesTemp)

	bytesTemp = make([]byte, nameLength)
	err = binary.Read(file, binary.BigEndian, bytesTemp)
	kl.DatasetName = string(bytesTemp)

	bytesTemp = make([]byte, 4)
	err = binary.Read(file, binary.BigEndian, bytesTemp)
	kl.NgramSize = binary.BigEndian.Uint32(bytesTemp)

	err = binary.Read(file, binary.BigEndian, bytesTemp)
	kl.SkipSize = binary.BigEndian.Uint32(bytesTemp)

	err = binary.Read(file, binary.BigEndian, bytesTemp)
	kl.NumNgrams = binary.BigEndian.Uint32(bytesTemp)

	bytesTemp = make([]byte, 1)
	err = binary.Read(file, binary.BigEndian, bytesTemp)
	if bytesTemp[0] == 0x00 {
		kl.CollectionMethod = NormalNgramming
	} else {
		if bytesTemp[0] == 0x01 {
			kl.CollectionMethod = HashGramming
		} else {
			kl.CollectionMethod = UnknownNgramMethod
		}
	}

	counter := uint32(0)
	for {
		bytesTemp = make([]byte, kl.NgramSize)
		err = binary.Read(file, binary.BigEndian, bytesTemp)
		kl.NGrams = append(kl.NGrams, bytesTemp)

		counter += 1
		if counter >= kl.NumNgrams {
			break
		}
	}

	return kl, nil
}

func ShowKeeplistInfo(filePath string) {
	kl, err := ParseKeepList(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read keeplist file: %v.\n", err)
	}

	kl.Print()
}

func KeepListCompare(filePathOne, filePathTwo string) {
	kl1, err := ParseKeepList(filePathOne)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read keeplist %s: %v.\n", filePathOne, err)
		return
	}

	kl2, err := ParseKeepList(filePathTwo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read keeplist %s: %v.\n", filePathTwo, err)
		return
	}

	fmt.Println("Dataset Names:")
	fmt.Printf("\t%s\n", kl1.DatasetName)
	fmt.Printf("\t%s\n", kl2.DatasetName)

	fmt.Println("Ngrams:")
	fmt.Printf("\t%dx %d-grams (%d skip)\n", kl1.NumNgrams, kl1.NgramSize, kl1.SkipSize)
	fmt.Printf("\t%dx %d-grams (%d skip)\n", kl2.NumNgrams, kl2.NgramSize, kl2.SkipSize)

	if kl1.NgramSize == kl2.NgramSize {
		sameGrams := 0
		minGramCount := math.Min(float64(kl1.NumNgrams), float64(kl2.NumNgrams))
		for _, gram1 := range kl1.NGrams {
			for _, gram2 := range kl2.NGrams {
				if bytes.Compare(gram1, gram2) == 0 {
					sameGrams += 1
					break
				}
			}
		}
		sameScore := (float64(sameGrams) / minGramCount) * 100.0
		fmt.Printf("Datasets have %1.2f%% ngrams in common.\n", sameScore)
	}
}
