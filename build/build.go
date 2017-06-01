package build

import "github.com/arduino/arduino-cli/opts"

type options struct {
	GetPrefs       []string `long:"get-pref" description:"Prints the value of the given preference to the standard output stream. When the value does not exist, nothing is printed.\nIf no preference is given as parameter, it prints all preferences."`
	Board          string   `long:"board" description:"Select the board to compile for."`
	InstallBoards  string   `long:"install-boards" description:"Fetches available board support (platform) list and install the specified one, along with its related tools."`
	InstallLibrary string   `long:"install-library" description:"Fetches available libraries list and install the specified one."`
}

/*
func (*options) Usage() string {
	return "[--board package:arch:board[:parameters]] [--pref name=value] [-v|--verbose] [--preserve-temp-files] <FILE.ino>"
}
*/

// Options for the build module
var Options options

func init() {
	opts.Parser.AddCommand("build", "Build a sketch", "", &Options)
}
