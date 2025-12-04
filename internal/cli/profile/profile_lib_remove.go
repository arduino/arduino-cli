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

func initLibRemoveCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	var destDir string
	var profileArg arguments.Profile
	removeCommand := &cobra.Command{
		Use:   fmt.Sprintf("remove %s[@%s]...", i18n.Tr("LIBRARY"), i18n.Tr("VERSION_NUMBER")),
		Short: i18n.Tr("Removes a library from a sketch profile."),
		Example: "" +
			"  " + os.Args[0] + " profile lib remove AudioZero -m my_profile\n" +
			"  " + os.Args[0] + " profile lib remove Arduino_JSON@0.2.0 --profile my_profile\n",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runLibRemoveCommand(cmd.Context(), srv, args, profileArg.Get(), destDir)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return arguments.GetInstallableLibs(cmd.Context(), srv), cobra.ShellCompDirectiveDefault
		},
	}
	profileArg.AddToCommand(removeCommand, srv)
	removeCommand.Flags().StringVar(&destDir, "dest-dir", "", i18n.Tr("Location of the sketch project file."))
	return removeCommand
}

func runLibRemoveCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer, args []string, profile, destDir string) {
	sketchPath := arguments.InitSketchPath(destDir)

	instance := instance.CreateAndInit(ctx, srv)
	libRefs, err := lib.ParseLibraryReferenceArgsAndAdjustCase(ctx, srv, instance, args)
	if err != nil {
		feedback.Fatal(i18n.Tr("Arguments error: %v", err), feedback.ErrBadArgument)
	}
	for _, lib := range libRefs {
		resp, err := srv.ProfileLibRemove(ctx, &rpc.ProfileLibRemoveRequest{
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
		})
		if err != nil {
			feedback.Fatal(fmt.Sprintf("%s: %v",
				i18n.Tr("Error removing library %[1]s from the profile", lib.Name), err), feedback.ErrGeneric)
		}
		feedback.PrintResult(libRemoveResult{
			RemovedLibraries: f.Map(resp.GetRemovedLibraries(), result.NewProfileLibraryReference),
			ProfileName:      resp.ProfileName})
	}
}

type libRemoveResult struct {
	RemovedLibraries []*result.ProfileLibraryReference `json:"removed_libraries"`
	ProfileName      string                            `json:"profile_name"`
}

func (lr libRemoveResult) Data() interface{} {
	return lr
}

func (lr libRemoveResult) String() string {
	res := fmt.Sprintln(i18n.Tr("The following libraries were removed from the profile %s:", lr.ProfileName))
	for _, lib := range lr.RemovedLibraries {
		res += fmt.Sprintf("  - %s\n", lib)
	}
	return res
}
