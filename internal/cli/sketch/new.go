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

	"github.com/arduino/arduino-cli/internal/arduino/globals"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initNewCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	var overwrite bool

	newCommand := &cobra.Command{
		Use:     "new",
		Short:   i18n.Tr("Create a new Sketch"),
		Long:    i18n.Tr("Create a new Sketch"),
		Example: "  " + os.Args[0] + " sketch new MultiBlinker",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runNewCommand(cmd.Context(), srv, args, overwrite)
		},
	}

	newCommand.Flags().BoolVarP(&overwrite, "overwrite", "f", false, i18n.Tr("Overwrites an existing .ino sketch."))

	return newCommand
}

func runNewCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer, args []string, overwrite bool) {
	logrus.Info("Executing `arduino-cli sketch new`")
	// Trim to avoid issues if user creates a sketch adding the .ino extesion to the name
	inputSketchName := args[0]
	trimmedSketchName := strings.TrimSuffix(inputSketchName, globals.MainFileValidExtension)

	var sketchDir string
	var sketchName string
	var sketchDirPath *paths.Path
	var err error

	if trimmedSketchName == "" {
		// `paths.New` returns nil with an empty string so `paths.Abs` panics.
		// if the name is empty we rely on the "new" command to fail nicely later
		// with the same logic in grpc and cli flows
		sketchDir = ""
		sketchName = ""
	} else {
		sketchDirPath, err = paths.New(trimmedSketchName).Abs()
		if err != nil {
			feedback.Fatal(i18n.Tr("Error creating sketch: %v", err), feedback.ErrGeneric)
		}
		sketchDir = sketchDirPath.Parent().String()
		sketchName = sketchDirPath.Base()
	}

	_, err = srv.NewSketch(ctx, &rpc.NewSketchRequest{
		SketchName: sketchName,
		SketchDir:  sketchDir,
		Overwrite:  overwrite,
	})
	if err != nil {
		feedback.Fatal(i18n.Tr("Error creating sketch: %v", err), feedback.ErrGeneric)
	}

	feedback.PrintResult(sketchResult{SketchDirPath: sketchDirPath})
}

type sketchResult struct {
	SketchDirPath *paths.Path `json:"sketch_path"`
}

func (ir sketchResult) Data() interface{} {
	return ir
}

func (ir sketchResult) String() string {
	return i18n.Tr("Sketch created in: %s", ir.SketchDirPath)
}
