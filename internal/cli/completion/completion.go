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

	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	completionNoDesc bool // Disable completion description for shells that support it
	tr               = i18n.Tr
)

// NewCommand created a new `completion` command
func NewCommand() *cobra.Command {
	completionCommand := &cobra.Command{
		Use:       "completion [bash|zsh|fish|powershell] [--no-descriptions]",
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.ExactArgs(1),
		Short:     i18n.Tr("Generates completion scripts"),
		Long:      i18n.Tr("Generates completion scripts for various shells"),
		Example: "  " + os.Args[0] + " completion bash > completion.sh\n" +
			"  " + "source completion.sh",
		Run: runCompletionCommand,
	}
	completionCommand.Flags().BoolVar(&completionNoDesc, "no-descriptions", false, i18n.Tr("Disable completion description for shells that support it"))

	return completionCommand
}

func runCompletionCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino-cli completion`")
	stdOut, _, err := feedback.DirectStreams()
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrGeneric)
	}
	if completionNoDesc && (args[0] == "powershell") {
		feedback.Fatal(i18n.Tr("Error: command description is not supported by %v", args[0]), feedback.ErrGeneric)
	}
	switch args[0] {
	case "bash":
		cmd.Root().GenBashCompletionV2(stdOut, !completionNoDesc)
	case "zsh":
		if completionNoDesc {
			cmd.Root().GenZshCompletionNoDesc(stdOut)
		} else {
			cmd.Root().GenZshCompletion(stdOut)
		}
	case "fish":
		cmd.Root().GenFishCompletion(stdOut, !completionNoDesc)
	case "powershell":
		cmd.Root().GenPowerShellCompletion(stdOut)
	}
}
