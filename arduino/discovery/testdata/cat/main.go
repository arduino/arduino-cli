// Echo stdin to stdout.
// This program is used for testing purposes, to make it available on all
// OS a tool equivalent to UNIX "cat".
package main

import (
	"io"
	"os"
)

func main() {
	io.Copy(os.Stdout, os.Stdin)
}
