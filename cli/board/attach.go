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

package board

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands/board"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initAttachCommand() *cobra.Command {
	attachCommand := &cobra.Command{
		Use:   "attach <port>|<FQBN> [sketchPath]",
		Short: "Attaches a sketch to a board.",
		Long:  "Attaches a sketch to a board.",
		Example: "  " + os.Args[0] + " board attach serial:///dev/ttyACM0\n" +
			"  " + os.Args[0] + " board attach serial:///dev/ttyACM0 HelloWorld\n" +
			"  " + os.Args[0] + " board attach arduino:samd:mkr1000",
		Args: cobra.RangeArgs(1, 2),
		Run:  runAttachCommand,
	}
	attachCommand.Flags().StringVar(&attachFlags.searchTimeout, "timeout", "5s",
		"The connected devices search timeout, raise it if your board doesn't show up (e.g. to 10s).")
	return attachCommand
}

var attachFlags struct {
	searchTimeout string // Expressed in a parsable duration, is the timeout for the list and attach commands.
}

func runAttachCommand(cmd *cobra.Command, args []string) {
	instance := instance.CreateAndInit()

	var path *paths.Path
	if len(args) > 1 {
		path = paths.New(args[1])
	} else {
		path = initSketchPath(path)
	}

	if _, err := board.Attach(context.Background(), &rpc.BoardAttachRequest{
		Instance:      instance,
		BoardUri:      args[0],
		SketchPath:    path.String(),
		SearchTimeout: attachFlags.searchTimeout,
	}, output.TaskProgress()); err != nil {
		feedback.Errorf("Attach board error: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
}

// initSketchPath returns the current working directory
func initSketchPath(sketchPath *paths.Path) *paths.Path {
	if sketchPath != nil {
		return sketchPath
	}

	wd, err := paths.Getwd()
	if err != nil {
		feedback.Errorf("Couldn't get current working directory: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
	logrus.Infof("Reading sketch from dir: %s", wd)
	return wd
}
