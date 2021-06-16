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
	"fmt"
	"os"
	"os/signal"
	"sort"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/debug"
	dbg "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/debug/v1"
	"github.com/arduino/arduino-cli/table"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/status"
)

var (
	fqbn        string
	port        string
	verbose     bool
	verify      bool
	interpreter string
	importDir   string
	printInfo   bool
	programmer  string
)

// NewCommand created a new `upload` command
func NewCommand() *cobra.Command {
	debugCommand := &cobra.Command{
		Use:     "debug",
		Short:   "Debug Arduino sketches.",
		Long:    "Debug Arduino sketches. (this command opens an interactive gdb session)",
		Example: "  " + os.Args[0] + " debug -b arduino:samd:mkr1000 -P atmel_ice /home/user/Arduino/MySketch",
		Args:    cobra.MaximumNArgs(1),
		Run:     run,
	}

	debugCommand.Flags().StringVarP(&fqbn, "fqbn", "b", "", "Fully Qualified Board Name, e.g.: arduino:avr:uno")
	debugCommand.Flags().StringVarP(&port, "port", "p", "", "Debug port, e.g.: COM10 or /dev/ttyACM0")
	debugCommand.Flags().StringVarP(&programmer, "programmer", "P", "", "Programmer to use for debugging")
	debugCommand.Flags().StringVar(&interpreter, "interpreter", "console", "Debug interpreter e.g.: console, mi, mi1, mi2, mi3")
	debugCommand.Flags().StringVarP(&importDir, "input-dir", "", "", "Directory containing binaries for debug.")
	debugCommand.Flags().BoolVarP(&printInfo, "info", "I", false, "Show metadata about the debug session instead of starting the debugger.")

	return debugCommand
}

func run(command *cobra.Command, args []string) {
	instance := instance.CreateAndInit()

	var path *paths.Path
	if len(args) > 0 {
		path = paths.New(args[0])
	}
	sketchPath := initSketchPath(path)

	debugConfigRequested := &dbg.DebugConfigRequest{
		Instance:    instance,
		Fqbn:        fqbn,
		SketchPath:  sketchPath.String(),
		Port:        port,
		Interpreter: interpreter,
		ImportDir:   importDir,
		Programmer:  programmer,
	}

	if printInfo {

		if res, err := debug.GetDebugConfig(context.Background(), debugConfigRequested); err != nil {
			if status, ok := status.FromError(err); ok {
				feedback.Errorf("Error getting Debug info: %v", status.Message())
				errorcodes.ExitWithGrpcStatus(status)
			}
			feedback.Errorf("Error getting Debug info: %v", err)
			os.Exit(errorcodes.ErrGeneric)
		} else {
			feedback.PrintResult(&debugInfoResult{res})
		}

	} else {

		// Intercept SIGINT and forward them to debug process
		ctrlc := make(chan os.Signal, 1)
		signal.Notify(ctrlc, os.Interrupt)

		if _, err := debug.Debug(context.Background(), debugConfigRequested, os.Stdin, os.Stdout, ctrlc); err != nil {
			feedback.Errorf("Error during Debug: %v", err)
			os.Exit(errorcodes.ErrGeneric)
		}

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

type debugInfoResult struct {
	info *dbg.GetDebugConfigResponse
}

func (r *debugInfoResult) Data() interface{} {
	return r.info
}

func (r *debugInfoResult) String() string {
	t := table.New()
	green := color.New(color.FgHiGreen)
	dimGreen := color.New(color.FgGreen)
	t.AddRow("Executable to debug", table.NewCell(r.info.GetExecutable(), green))
	t.AddRow("Toolchain type", table.NewCell(r.info.GetToolchain(), green))
	t.AddRow("Toolchain path", table.NewCell(r.info.GetToolchainPath(), dimGreen))
	t.AddRow("Toolchain prefix", table.NewCell(r.info.GetToolchainPrefix(), dimGreen))
	if len(r.info.GetToolchainConfiguration()) > 0 {
		conf := properties.NewFromHashmap(r.info.GetToolchainConfiguration())
		keys := conf.Keys()
		sort.Strings(keys)
		t.AddRow("Toolchain custom configurations")
		for _, k := range keys {
			t.AddRow(table.NewCell(" - "+k, dimGreen), table.NewCell(conf.Get(k), dimGreen))
		}
	}
	t.AddRow("GDB Server type", table.NewCell(r.info.GetServer(), green))
	t.AddRow("GDB Server path", table.NewCell(r.info.GetServerPath(), dimGreen))
	if len(r.info.GetServerConfiguration()) > 0 {
		conf := properties.NewFromHashmap(r.info.GetServerConfiguration())
		keys := conf.Keys()
		sort.Strings(keys)
		t.AddRow(fmt.Sprintf("%s custom configurations", r.info.GetServer()))
		for _, k := range keys {
			t.AddRow(table.NewCell(" - "+k, dimGreen), table.NewCell(conf.Get(k), dimGreen))
		}
	}
	return t.Render()
}
