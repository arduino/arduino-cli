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
	"sort"

	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/feedback/result"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initDepsCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	var noOverwrite bool
	depsCommand := &cobra.Command{
		Use:   fmt.Sprintf("deps %s[@%s]...", i18n.Tr("LIBRARY"), i18n.Tr("VERSION_NUMBER")),
		Short: i18n.Tr("Check dependencies status for the specified library."),
		Long:  i18n.Tr("Check dependencies status for the specified library."),
		Example: "" +
			"  " + os.Args[0] + " lib deps AudioZero       # " + i18n.Tr("for the latest version.") + "\n" +
			"  " + os.Args[0] + " lib deps AudioZero@1.0.0 # " + i18n.Tr("for the specific version."),
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runDepsCommand(cmd.Context(), srv, args, noOverwrite)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return arguments.GetInstalledLibraries(cmd.Context(), srv), cobra.ShellCompDirectiveDefault
		},
	}
	depsCommand.Flags().BoolVar(&noOverwrite, "no-overwrite", false, i18n.Tr("Do not try to update library dependencies if already installed."))
	return depsCommand
}

func runDepsCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer, args []string, noOverwrite bool) {
	instance := instance.CreateAndInit(ctx, srv)

	logrus.Info("Executing `arduino-cli lib deps`")
	libRef, err := ParseLibraryReferenceArgAndAdjustCase(ctx, srv, instance, args[0])
	if err != nil {
		feedback.Fatal(i18n.Tr("Arguments error: %v", err), feedback.ErrBadArgument)
	}

	deps, err := srv.LibraryResolveDependencies(ctx, &rpc.LibraryResolveDependenciesRequest{
		Instance:                      instance,
		Name:                          libRef.Name,
		Version:                       libRef.Version,
		DoNotUpdateInstalledLibraries: noOverwrite,
	})
	if err != nil {
		feedback.Fatal(i18n.Tr("Error resolving dependencies for %[1]s: %[2]s", libRef, err), feedback.ErrGeneric)
	}

	feedback.PrintResult(&checkDepResult{deps: result.NewLibraryResolveDependenciesResponse(deps)})
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type checkDepResult struct {
	deps *result.LibraryResolveDependenciesResponse
}

func (dr checkDepResult) Data() interface{} {
	return dr.deps
}

func (dr checkDepResult) String() string {
	if dr.deps == nil || dr.deps.Dependencies == nil {
		return ""
	}
	res := ""
	deps := dr.deps.Dependencies

	// Sort dependencies alphabetically and then puts installed ones on top
	sort.Slice(deps, func(i, j int) bool {
		return deps[i].Name < deps[j].Name
	})
	sort.SliceStable(deps, func(i, j int) bool {
		return deps[i].VersionInstalled != "" && deps[j].VersionInstalled == ""
	})

	for _, dep := range deps {
		if dep == nil {
			continue
		}
		res += outputDep(dep)
	}
	return res
}

func outputDep(dep *result.LibraryDependencyStatus) string {
	res := ""
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)
	switch dep.VersionInstalled {
	case "":
		res += i18n.Tr("%s must be installed.",
			red.Sprintf("✕ %s %s", dep.Name, dep.VersionRequired))
	case dep.VersionRequired:
		res += i18n.Tr("%s is already installed.",
			green.Sprintf("✓ %s %s", dep.Name, dep.VersionRequired))
	default:
		res += i18n.Tr("%[1]s is required but %[2]s is currently installed.",
			yellow.Sprintf("✕ %s %s", dep.Name, dep.VersionRequired),
			yellow.Sprintf("%s", dep.VersionInstalled))
	}
	res += "\n"
	return res
}
