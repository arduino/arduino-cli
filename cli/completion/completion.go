// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package completion

import (
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/spf13/cobra"
)

var (
	completionNoDesc bool //Disable completion description for shells that support it
)

// NewCommand created a new `completion` command
func NewCommand() *cobra.Command {
	command := &cobra.Command{
		Use:       "completion [bash|zsh|fish] [--no-descriptions]",
		ValidArgs: []string{"bash", "zsh", "fish"},
		Args:      cobra.ExactArgs(1),
		Short:     "Generates completion scripts",
		Long:      "Generates completion scripts for various shells",
		Example: "  " + os.Args[0] + " completion bash > completion.sh\n" +
			"  " + "source completion.sh",
		Run: run,
	}
	command.Flags().BoolVar(&completionNoDesc, "no-descriptions", false, "Disable completion description for shells that support it")

	return command
}

func run(cmd *cobra.Command, args []string) {
	if completionNoDesc && (args[0] == "bash") {
		feedback.Errorf("Error: command description is not supported by %v", args[0])
		os.Exit(errorcodes.ErrGeneric)
	}
	switch args[0] {
	case "bash":
		cmd.Root().GenBashCompletion(os.Stdout)
		break
	case "zsh":
		if completionNoDesc {
			cmd.Root().GenZshCompletionNoDesc(os.Stdout)
		} else {
			cmd.Root().GenZshCompletion(os.Stdout)
		}
		break
	case "fish":
		cmd.Root().GenFishCompletion(os.Stdout, !completionNoDesc)
		break
	}
}
