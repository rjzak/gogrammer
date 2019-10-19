package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

func ShowKeeplistInfo(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open %s for reading: %v.\n", filePath, err)
		os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "%s is not a keeplist.\n", filePath)
		return
	}

	bytesTemp = make([]byte, 4)
	err = binary.Read(file, binary.BigEndian, bytesTemp)
	nameLength := binary.BigEndian.Uint32(bytesTemp)

	bytesTemp = make([]byte, nameLength)
	err = binary.Read(file, binary.BigEndian, bytesTemp)
	fmt.Printf("Dataset name: %s\n", string(bytesTemp))

	bytesTemp = make([]byte, 4)
	err = binary.Read(file, binary.BigEndian, bytesTemp)
	fmt.Printf("Ngram size: %d\n", binary.BigEndian.Uint32(bytesTemp))

	err = binary.Read(file, binary.BigEndian, bytesTemp)
	fmt.Printf("Skip size: %d\n", binary.BigEndian.Uint32(bytesTemp))

	err = binary.Read(file, binary.BigEndian, bytesTemp)
	fmt.Printf("Number of n-grams: %d\n", binary.BigEndian.Uint32(bytesTemp))
}
