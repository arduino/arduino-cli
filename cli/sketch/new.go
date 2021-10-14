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

package sketch

import (
	"os"
	"strings"

	"github.com/arduino/arduino-cli/arduino/globals"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	sk "github.com/arduino/arduino-cli/commands/sketch"
	paths "github.com/arduino/go-paths-helper"
	"github.com/spf13/cobra"
)

func initNewCommand() *cobra.Command {
	newCommand := &cobra.Command{
		Use:     "new",
		Short:   tr("Create a new Sketch"),
		Long:    tr("Create a new Sketch"),
		Example: "  " + os.Args[0] + " sketch new MultiBlinker",
		Args:    cobra.ExactArgs(1),
		Run:     runNewCommand,
	}
	return newCommand
}

func runNewCommand(cmd *cobra.Command, args []string) {
	// Trim to avoid issues if user creates a sketch adding the .ino extesion to the name
	sketchName := args[0]
	trimmedSketchName := strings.TrimSuffix(sketchName, globals.MainFileValidExtension)
	sketchDirPath, err := paths.New(trimmedSketchName).Abs()
	if err != nil {
		feedback.Errorf(tr("Error creating sketch: %v"), err)
		os.Exit(errorcodes.ErrGeneric)
	}
	_, err = sk.CreateSketch(sketchDirPath)
	if err != nil {
		feedback.Errorf(tr("Error creating sketch: %v"), err)
		os.Exit(errorcodes.ErrGeneric)
	}

	feedback.Print(tr("Sketch created in: %s", sketchDirPath.String()))
}
