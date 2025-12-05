// This file is part of arduino-cli.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
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

package profile

import (
	"context"
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/feedback/result"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/cli/lib"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/spf13/cobra"
	"go.bug.st/f"
)

func initLibAddCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	var sketchDir string
	var noDeps bool
	var noOverwrite bool
	var profileArg arguments.Profile
	addCommand := &cobra.Command{
		Use:   fmt.Sprintf("add %s[@%s]...", i18n.Tr("LIBRARY"), i18n.Tr("VERSION_NUMBER")),
		Short: i18n.Tr("Adds a library to a sketch profile."),
		Example: "" +
			"  " + os.Args[0] + " profile lib add AudioZero -m my_profile\n" +
			"  " + os.Args[0] + " profile lib add Arduino_JSON@0.2.0 --profile my_profile\n",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runLibAddCommand(cmd.Context(), args, srv, profileArg.Get(), sketchDir, noDeps, noOverwrite)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return arguments.GetInstallableLibs(cmd.Context(), srv), cobra.ShellCompDirectiveDefault
		},
	}
	profileArg.AddToCommand(addCommand, srv)
	addCommand.Flags().StringVar(&sketchDir, "sketch-path", "", i18n.Tr("Location of the sketch."))
	addCommand.Flags().BoolVar(&noDeps, "no-deps", false, i18n.Tr("Do not add dependencies."))
	addCommand.Flags().BoolVar(&noOverwrite, "no-overwrite", false, i18n.Tr("Do not overwrite already added libraries."))
	return addCommand
}

func runLibAddCommand(ctx context.Context, args []string, srv rpc.ArduinoCoreServiceServer, profile, sketchDir string, noAddDeps, noOverwrite bool) {
	sketchPath := arguments.InitSketchPath(sketchDir)

	instance := instance.CreateAndInit(ctx, srv)
	libRefs, err := lib.ParseLibraryReferenceArgsAndAdjustCase(ctx, srv, instance, args)
	if err != nil {
		feedback.Fatal(i18n.Tr("Arguments error: %v", err), feedback.ErrBadArgument)
	}
	addDeps := !noAddDeps
	for _, lib := range libRefs {
		resp, err := srv.ProfileLibAdd(ctx, &rpc.ProfileLibAddRequest{
			Instance:    instance,
			SketchPath:  sketchPath.String(),
			ProfileName: profile,
			Library: &rpc.ProfileLibraryReference{
				Library: &rpc.ProfileLibraryReference_IndexLibrary_{
					IndexLibrary: &rpc.ProfileLibraryReference_IndexLibrary{
						Name:    lib.Name,
						Version: lib.Version,
					},
				},
			},
			AddDependencies: &addDeps,
			NoOverwrite:     &noOverwrite,
		})
		if err != nil {
			feedback.Fatal(i18n.Tr("Error adding %s: %v", lib.Name, err), feedback.ErrGeneric)
		}
		feedback.PrintResult(libAddResult{
			AddedLibraries:   f.Map(resp.GetAddedLibraries(), result.NewProfileLibraryReference),
			SkippedLibraries: f.Map(resp.GetSkippedLibraries(), result.NewProfileLibraryReference),
			ProfileName:      resp.ProfileName,
		})
	}
}

type libAddResult struct {
	AddedLibraries   []*result.ProfileLibraryReference `json:"added_libraries"`
	SkippedLibraries []*result.ProfileLibraryReference `json:"skipped_libraries"`
	ProfileName      string                            `json:"profile_name"`
}

func (lr libAddResult) Data() any {
	return lr
}

func (lr libAddResult) String() string {
	res := ""
	if len(lr.AddedLibraries) > 0 {
		res += fmt.Sprintln(i18n.Tr("The following libraries were added to the profile %s:", lr.ProfileName))
		for _, l := range lr.AddedLibraries {
			res += fmt.Sprintf("  - %s\n", l)
		}
	}
	if len(lr.SkippedLibraries) > 0 {
		res += fmt.Sprintln(i18n.Tr("The following libraries were already present in the profile %s and were not modified:", lr.ProfileName))
		for _, l := range lr.SkippedLibraries {
			res += fmt.Sprintf("  - %s\n", l)
		}
	}
	return res
}
