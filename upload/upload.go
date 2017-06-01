package upload

import (
	"github.com/arduino/arduino-cli/opts"
)

// Options for the command line parser
type options struct {
	Board string `long:"board" description:"Select the board to compile for." optional:"yes"`
}

/*
func (*Options) Usage() string {
	return "[--board package:arch:board[:parameters]] [--pref name=value] [-v|--verbose] [--preserve-temp-files] <FILE.ino>"
}
*/

// Options for the build module
var Options options

func init() {
	opts.Parser.AddCommand("upload", "Upload a sketch", "", &Options)
}
