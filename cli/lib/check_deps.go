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

package lib

import (
	"context"
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/lib"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func initDepsCommand() *cobra.Command {
	depsCommand := &cobra.Command{
		Use:   "deps LIBRARY[@VERSION_NUMBER](S)",
		Short: "Check dependencies status for the specified library.",
		Long:  "Check dependencies status for the specified library.",
		Example: "" +
			"  " + os.Args[0] + " lib deps AudioZero       # for the latest version.\n" +
			"  " + os.Args[0] + " lib deps AudioZero@1.0.0 # for the specific version.",
		Args: cobra.ExactArgs(1),
		Run:  runDepsCommand,
	}
	return depsCommand
}

func runDepsCommand(cmd *cobra.Command, args []string) {
	instance := instance.CreateAndInit()
	libRef, err := ParseLibraryReferenceArgAndAdjustCase(instance, args[0])
	if err != nil {
		feedback.Errorf("Arguments error: %v", err)
		os.Exit(errorcodes.ErrBadArgument)
	}

	deps, err := lib.LibraryResolveDependencies(context.Background(), &rpc.LibraryResolveDependenciesRequest{
		Instance: instance,
		Name:     libRef.Name,
		Version:  libRef.Version,
	})
	if err != nil {
		feedback.Errorf("Error resolving dependencies for %s: %s", libRef, err)
	}

	feedback.PrintResult(&checkDepResult{deps: deps})
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type checkDepResult struct {
	deps *rpc.LibraryResolveDependenciesResponse
}

func (dr checkDepResult) Data() interface{} {
	return dr.deps
}

func (dr checkDepResult) String() string {
	res := ""
	for _, dep := range dr.deps.GetDependencies() {
		res += outputDep(dep)
	}
	return res
}

func outputDep(dep *rpc.LibraryDependencyStatus) string {
	res := ""
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)
	if dep.GetVersionInstalled() == "" {
		res += fmt.Sprintf("%s must be installed.\n",
			red.Sprintf("✕ %s %s", dep.GetName(), dep.GetVersionRequired()))
	} else if dep.GetVersionInstalled() == dep.GetVersionRequired() {
		res += fmt.Sprintf("%s is already installed.\n",
			green.Sprintf("✓ %s %s", dep.GetName(), dep.GetVersionRequired()))
	} else {
		res += fmt.Sprintf("%s is required but %s is currently installed.\n",
			yellow.Sprintf("✕ %s %s", dep.GetName(), dep.GetVersionRequired()),
			yellow.Sprintf("%s", dep.GetVersionInstalled()))
	}
	return res
}
