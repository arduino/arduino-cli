// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
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

package debug

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/commands/debug"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/feedback/result"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newDebugCheckCommand() *cobra.Command {
	var (
		fqbnArg     arguments.Fqbn
		portArgs    arguments.Port
		interpreter string
		programmer  arguments.Programmer
	)
	debugCheckCommand := &cobra.Command{
		Use:     "check",
		Short:   tr("Check if the given board/programmer combination supports debugging."),
		Example: "  " + os.Args[0] + " debug check -b arduino:samd:mkr1000 -P atmel_ice",
		Run: func(cmd *cobra.Command, args []string) {
			runDebugCheckCommand(&portArgs, &fqbnArg, interpreter, &programmer)
		},
	}
	fqbnArg.AddToCommand(debugCheckCommand)
	portArgs.AddToCommand(debugCheckCommand)
	programmer.AddToCommand(debugCheckCommand)
	debugCheckCommand.Flags().StringVar(&interpreter, "interpreter", "console", tr("Debug interpreter e.g.: %s", "console, mi, mi1, mi2, mi3"))
	return debugCheckCommand
}

func runDebugCheckCommand(portArgs *arguments.Port, fqbnArg *arguments.Fqbn, interpreter string, programmerArg *arguments.Programmer) {
	instance := instance.CreateAndInit()
	logrus.Info("Executing `arduino-cli debug`")

	port, err := portArgs.GetPort(instance, "", "")
	if err != nil {
		feedback.FatalError(err, feedback.ErrBadArgument)
	}
	fqbn := fqbnArg.String()
	resp, err := debug.IsDebugSupported(context.Background(), &rpc.IsDebugSupportedRequest{
		Instance:    instance,
		Fqbn:        fqbn,
		Port:        port,
		Interpreter: interpreter,
		Programmer:  programmerArg.String(instance, fqbn),
	})
	if err != nil {
		feedback.FatalError(err, feedback.ErrGeneric)
	}
	feedback.PrintResult(&debugCheckResult{result.NewIsDebugSupportedResponse(resp)})
}

type debugCheckResult struct {
	Result *result.IsDebugSupportedResponse
}

func (d *debugCheckResult) Data() interface{} {
	return d.Result
}

func (d *debugCheckResult) String() string {
	if d.Result.Supported {
		return tr("The given board/programmer configuration supports debugging.")
	}
	return tr("The given board/programmer configuration does NOT support debugging.")
}
