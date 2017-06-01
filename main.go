package main

import (
	"fmt"
	"os"

	_ "github.com/arduino/arduino-cli/build"
	_ "github.com/arduino/arduino-cli/libraries"
	"github.com/arduino/arduino-cli/opts"
	_ "github.com/arduino/arduino-cli/upload"
)

const version = "0.1.0"

type versionOpts struct {
}

func (*versionOpts) Execute(args []string) error {
	fmt.Println("arduino-cli " + version)
	return nil
}

func main() {
	opts.Parser.AddCommand("version", "Prints version info", "", &versionOpts{})
	args, err := opts.Parse()
	if err != nil {
		os.Exit(1)
	}

	fmt.Println("remaining args: ", args)
}
