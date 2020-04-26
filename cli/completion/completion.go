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
	"bytes"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	completionNoDesc bool //Disable completion description for shells that support it
)

// NewCommand created a new `version` command
func NewCommand() *cobra.Command {
	command := &cobra.Command{
		Use:       "completion [bash|zsh|fish] [--no-descriptions]",
		ValidArgs: []string{"bash\t", "zsh", "fish"},
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
	switch args[0] {
	case "bash":
		cmd.Root().GenBashCompletion(os.Stdout)
		break
	case "zsh":
		cmd.Root().GenZshCompletion(os.Stdout)
		break
	case "fish":
		buf := new(bytes.Buffer)
		cmd.Root().GenFishCompletion(buf, !completionNoDesc)
		newstring := strings.ReplaceAll(buf.String(), "arduino-cli_comp", "arduino_cli_comp") //required because fish does not support env variables with "-" in the name
		os.Stdout.WriteString(newstring)
		break
	}
}
