package main

import (
	"fmt"
	"os"
)

func main() {
	for _, env := range os.Environ() {
		fmt.Fprintln(os.Stderr, "ENV>", env)
	}
}
