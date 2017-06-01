package opts

import "github.com/jessevdk/go-flags"

// GeneralOpts is a struct that defines the command line options that are generic for the application
type GeneralOpts struct {
	// Slice of bool will append 'true' each time the option is encountered (can be set multiple times, like -vvv)
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information."`
}

// General contains the current settings for the generic command line options
var General GeneralOpts

// Parser is the main parser for the command line options
var Parser *flags.Parser

func init() {
	Parser = flags.NewParser(&General, flags.Default)
}

// Parse perform command line parsing
func Parse() ([]string, error) {
	args, err := Parser.Parse()
	return args, err
}
