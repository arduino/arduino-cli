package iohelpers

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// ReadBinaryFromFile reads an array of bytes from a file. Panics if an error occurs.
func ReadBinaryFromFile(file string) []byte {
	key, err := ioutil.ReadFile(file)
	if err != nil {
		panic(fmt.Sprintf("Failed to load %q from folder: %v", file, err))
	}
	return key
}

// ReadFirstLineOfFile reads and returns the first line of a file as a string.
func ReadFirstLineOfFile(file string) string {
	return strings.TrimSpace(strings.Split(string(ReadBinaryFromFile(file)), "\n")[0])
}
