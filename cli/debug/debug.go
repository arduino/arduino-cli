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

package debug

import (
	"context"
	"os"
	"os/signal"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/debug"
	dbg "github.com/arduino/arduino-cli/rpc/debug"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	fqbn       string
	port       string
	verbose    bool
	verify     bool
	importFile string
)

// NewCommand created a new `upload` command
func NewCommand() *cobra.Command {
	debugCommand := &cobra.Command{
		Use:     "debug",
		Short:   "Debug Arduino sketches.",
		Long:    "Debug Arduino sketches. (this command opens an interactive gdb session)",
		Example: "  " + os.Args[0] + " debug -b arduino:samd:mkr1000  /home/user/Arduino/MySketch",
		Args:    cobra.MaximumNArgs(1),
		Run:     run,
	}

	debugCommand.Flags().StringVarP(&fqbn, "fqbn", "b", "", "Fully Qualified Board Name, e.g.: arduino:avr:uno")
	debugCommand.Flags().StringVarP(&port, "port", "p", "", "Upload port, e.g.: COM10 or /dev/ttyACM0")
	debugCommand.Flags().StringVarP(&importFile, "input", "i", "", "Input file to be uploaded for debug.")

	return debugCommand
}

func run(command *cobra.Command, args []string) {
	instance, err := instance.CreateInstance()
	if err != nil {
		feedback.Errorf("Error during Debug: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	var path *paths.Path
	if len(args) > 0 {
		path = paths.New(args[0])
	}
	sketchPath := initSketchPath(path)

	// Intercept SIGINT and forward them to debug process
	ctrlc := make(chan os.Signal, 1)
	signal.Notify(ctrlc, os.Interrupt)

	if _, err := debug.Debug(context.Background(), &dbg.DebugConfigReq{
		Instance:   &dbg.Instance{Id: instance.GetId()},
		Fqbn:       fqbn,
		SketchPath: sketchPath.String(),
		Port:       port,
		ImportFile: importFile,
	}, os.Stdin, os.Stdout, ctrlc); err != nil {
		feedback.Errorf("Error during Debug: %v", err)
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
