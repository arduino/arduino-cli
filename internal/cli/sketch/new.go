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
	"context"
	"os"
	"strings"
	"regexp"
	"fmt"
	"github.com/arduino/arduino-cli/arduino/globals"
	sk "github.com/arduino/arduino-cli/commands/sketch"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initNewCommand() *cobra.Command {
	var overwrite bool

	newCommand := &cobra.Command{
		Use:     "new",
		Short:   tr("Create a new Sketch"),
		Long:    tr("Create a new Sketch"),
		Example: "  " + os.Args[0] + " sketch new MultiBlinker",
		Args:    cobra.ExactArgs(1),
		Run:     func(cmd *cobra.Command, args []string) {
			re := regexp.MustCompile("^[a-zA-Z].")
			if !re.MatchString(args[0]) {
				fmt.Println("Error: Value can only contain alphabetic characters")
				return
			}
			runNewCommand(args, overwrite)
		},
	}

	newCommand.Flags().BoolVarP(&overwrite, "overwrite", "f", false, tr("Overwrites an existing .ino sketch."))

	return newCommand
}

func runNewCommand(args []string, overwrite bool) {
	logrus.Info("Executing `arduino-cli sketch new`")
	// Trim to avoid issues if user creates a sketch adding the .ino extesion to the name
	sketchName := args[0]
	trimmedSketchName := strings.TrimSuffix(sketchName, globals.MainFileValidExtension)
	sketchDirPath, err := paths.New(trimmedSketchName).Abs()
	if err != nil {
		feedback.Fatal(tr("Error creating sketch: %v", err), feedback.ErrGeneric)
	}
	_, err = sk.NewSketch(context.Background(), &rpc.NewSketchRequest{
		Instance:   nil,
		SketchName: sketchDirPath.Base(),
		SketchDir:  sketchDirPath.Parent().String(),
		Overwrite:  overwrite,
	})
	if err != nil {
		feedback.Fatal(tr("Error creating sketch: %v", err), feedback.ErrGeneric)
	}

	feedback.Print(tr("Sketch created in: %s", sketchDirPath))
}
