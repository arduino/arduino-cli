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

	"github.com/arduino/arduino-cli/cli/arguments"
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
		Use:   fmt.Sprintf("deps %s[@%s]...", tr("LIBRARY"), tr("VERSION_NUMBER")),
		Short: tr("Check dependencies status for the specified library."),
		Long:  tr("Check dependencies status for the specified library."),
		Example: "" +
			"  " + os.Args[0] + " lib deps AudioZero       # " + tr("for the latest version.") + "\n" +
			"  " + os.Args[0] + " lib deps AudioZero@1.0.0 # " + tr("for the specific version."),
		Args: cobra.ExactArgs(1),
		Run:  runDepsCommand,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return arguments.GetUninstallableLibs(), cobra.ShellCompDirectiveDefault
		},
	}
	return depsCommand
}

func runDepsCommand(cmd *cobra.Command, args []string) {
	instance := instance.CreateAndInit()
	libRef, err := ParseLibraryReferenceArgAndAdjustCase(instance, args[0])
	if err != nil {
		feedback.Errorf(tr("Arguments error: %v"), err)
		os.Exit(errorcodes.ErrBadArgument)
	}

	deps, err := lib.LibraryResolveDependencies(context.Background(), &rpc.LibraryResolveDependenciesRequest{
		Instance: instance,
		Name:     libRef.Name,
		Version:  libRef.Version,
	})
	if err != nil {
		feedback.Errorf(tr("Error resolving dependencies for %[1]s: %[2]s", libRef, err))
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
		res += tr("%s must be installed.",
			red.Sprintf("✕ %s %s", dep.GetName(), dep.GetVersionRequired()))
	} else if dep.GetVersionInstalled() == dep.GetVersionRequired() {
		res += tr("%s is already installed.",
			green.Sprintf("✓ %s %s", dep.GetName(), dep.GetVersionRequired()))
	} else {
		res += tr("%[1]s is required but %[2]s is currently installed.",
			yellow.Sprintf("✕ %s %s", dep.GetName(), dep.GetVersionRequired()),
			yellow.Sprintf("%s", dep.GetVersionInstalled()))
	}
	res += "\n"
	return res
}
