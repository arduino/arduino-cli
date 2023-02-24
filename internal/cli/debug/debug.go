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
	"sort"

	"github.com/arduino/arduino-cli/commands/debug"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	dbg "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/debug/v1"
	"github.com/arduino/arduino-cli/table"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	fqbnArg     arguments.Fqbn
	portArgs    arguments.Port
	interpreter string
	importDir   string
	printInfo   bool
	programmer  arguments.Programmer
	tr          = i18n.Tr
)

// NewCommand created a new `upload` command
func NewCommand() *cobra.Command {
	debugCommand := &cobra.Command{
		Use:     "debug",
		Short:   tr("Debug Arduino sketches."),
		Long:    tr("Debug Arduino sketches. (this command opens an interactive gdb session)"),
		Example: "  " + os.Args[0] + " debug -b arduino:samd:mkr1000 -P atmel_ice /home/user/Arduino/MySketch",
		Args:    cobra.MaximumNArgs(1),
		Run:     runDebugCommand,
	}

	fqbnArg.AddToCommand(debugCommand)
	portArgs.AddToCommand(debugCommand)
	programmer.AddToCommand(debugCommand)
	debugCommand.Flags().StringVar(&interpreter, "interpreter", "console", tr("Debug interpreter e.g.: %s", "console, mi, mi1, mi2, mi3"))
	debugCommand.Flags().StringVarP(&importDir, "input-dir", "", "", tr("Directory containing binaries for debug."))
	debugCommand.Flags().BoolVarP(&printInfo, "info", "I", false, tr("Show metadata about the debug session instead of starting the debugger."))

	return debugCommand
}

func runDebugCommand(command *cobra.Command, args []string) {
	instance := instance.CreateAndInit()
	logrus.Info("Executing `arduino-cli debug`")

	path := ""
	if len(args) > 0 {
		path = args[0]
	}

	sketchPath := arguments.InitSketchPath(path)
	sk := arguments.MustNewSketch(sketchPath)
	fqbn, port := arguments.CalculateFQBNAndPort(&portArgs, &fqbnArg, instance, sk)
	debugConfigRequested := &dbg.DebugConfigRequest{
		Instance:    instance,
		Fqbn:        fqbn,
		SketchPath:  sketchPath.String(),
		Port:        port,
		Interpreter: interpreter,
		ImportDir:   importDir,
		Programmer:  programmer.String(),
	}

	if printInfo {

		if res, err := debug.GetDebugConfig(context.Background(), debugConfigRequested); err != nil {
			feedback.Fatal(tr("Error getting Debug info: %v", err), feedback.ErrBadArgument)
		} else {
			feedback.PrintResult(&debugInfoResult{res})
		}

	} else {

		// Intercept SIGINT and forward them to debug process
		ctrlc := make(chan os.Signal, 1)
		signal.Notify(ctrlc, os.Interrupt)

		in, out, err := feedback.InteractiveStreams()
		if err != nil {
			feedback.FatalError(err, feedback.ErrBadArgument)
		}
		if _, err := debug.Debug(context.Background(), debugConfigRequested, in, out, ctrlc); err != nil {
			feedback.Fatal(tr("Error during Debug: %v", err), feedback.ErrGeneric)
		}

	}
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
	t.AddRow(tr("Executable to debug"), table.NewCell(r.info.GetExecutable(), green))
	t.AddRow(tr("Toolchain type"), table.NewCell(r.info.GetToolchain(), green))
	t.AddRow(tr("Toolchain path"), table.NewCell(r.info.GetToolchainPath(), dimGreen))
	t.AddRow(tr("Toolchain prefix"), table.NewCell(r.info.GetToolchainPrefix(), dimGreen))
	if len(r.info.GetToolchainConfiguration()) > 0 {
		conf := properties.NewFromHashmap(r.info.GetToolchainConfiguration())
		keys := conf.Keys()
		sort.Strings(keys)
		t.AddRow(tr("Toolchain custom configurations"))
		for _, k := range keys {
			t.AddRow(table.NewCell(" - "+k, dimGreen), table.NewCell(conf.Get(k), dimGreen))
		}
	}
	t.AddRow(tr("GDB Server type"), table.NewCell(r.info.GetServer(), green))
	t.AddRow(tr("GDB Server path"), table.NewCell(r.info.GetServerPath(), dimGreen))
	if len(r.info.GetServerConfiguration()) > 0 {
		conf := properties.NewFromHashmap(r.info.GetServerConfiguration())
		keys := conf.Keys()
		sort.Strings(keys)
		t.AddRow(tr("Configuration options for %s", r.info.GetServer()))
		for _, k := range keys {
			t.AddRow(table.NewCell(" - "+k, dimGreen), table.NewCell(conf.Get(k), dimGreen))
		}
	}
	return t.Render()
}
