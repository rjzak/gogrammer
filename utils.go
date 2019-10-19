package main

import "os"

var MAGIC = [8]byte{0x47, 0x4F, 0x5F, 0x47, 0x72, 0x61, 0x6D, 0x73}

func exists(path string) (bool) {
	_, err := os.Stat(path)
	if err == nil { return true }
	if os.IsNotExist(err) { return false }
	return true
}
