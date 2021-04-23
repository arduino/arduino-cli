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

package upload

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/arduino/sketches"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/upload"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	fqbn       string
	port       string
	verbose    bool
	verify     bool
	importDir  string
	importFile string
	programmer string
)

// NewCommand created a new `upload` command
func NewCommand() *cobra.Command {
	uploadCommand := &cobra.Command{
		Use:     "upload",
		Short:   "Upload Arduino sketches.",
		Long:    "Upload Arduino sketches. This does NOT compile the sketch prior to upload.",
		Example: "  " + os.Args[0] + " upload /home/user/Arduino/MySketch",
		Args:    cobra.MaximumNArgs(1),
		PreRun:  checkFlagsConflicts,
		Run:     run,
	}

	uploadCommand.Flags().StringVarP(&fqbn, "fqbn", "b", "", "Fully Qualified Board Name, e.g.: arduino:avr:uno")
	uploadCommand.Flags().StringVarP(&port, "port", "p", "", "Upload port, e.g.: COM10 or /dev/ttyACM0")
	uploadCommand.Flags().StringVarP(&importDir, "input-dir", "", "", "Directory containing binaries to upload.")
	uploadCommand.Flags().StringVarP(&importFile, "input-file", "i", "", "Binary file to upload.")
	uploadCommand.Flags().BoolVarP(&verify, "verify", "t", false, "Verify uploaded binary after the upload.")
	uploadCommand.Flags().BoolVarP(&verbose, "verbose", "v", false, "Optional, turns on verbose mode.")
	uploadCommand.Flags().StringVarP(&programmer, "programmer", "P", "", "Optional, use the specified programmer to upload.")

	return uploadCommand
}

func checkFlagsConflicts(command *cobra.Command, args []string) {
	if importFile != "" && importDir != "" {
		feedback.Errorf("error: --input-file and --input-dir flags cannot be used together")
		os.Exit(errorcodes.ErrBadArgument)
	}
}

func run(command *cobra.Command, args []string) {
	instance := instance.CreateAndInit()

	var path *paths.Path
	if len(args) > 0 {
		path = paths.New(args[0])
	}
	sketchPath := initSketchPath(path)

	// .pde files are still supported but deprecated, this warning urges the user to rename them
	if files := sketches.CheckForPdeFiles(sketchPath); len(files) > 0 {
		feedback.Error("Sketches with .pde extension are deprecated, please rename the following files to .ino:")
		for _, f := range files {
			feedback.Error(f)
		}
	}

	if _, err := upload.Upload(context.Background(), &rpc.UploadRequest{
		Instance:   instance,
		Fqbn:       fqbn,
		SketchPath: sketchPath.String(),
		Port:       port,
		Verbose:    verbose,
		Verify:     verify,
		ImportFile: importFile,
		ImportDir:  importDir,
		Programmer: programmer,
	}, os.Stdout, os.Stderr); err != nil {
		feedback.Errorf("Error during Upload: %v", err)
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
