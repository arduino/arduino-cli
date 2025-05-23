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
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/cli/lib"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/spf13/cobra"
)

func initLibCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	libCommand := &cobra.Command{
		Use:   "lib",
		Short: i18n.Tr("Profile commands about libraries."),
		Long:  i18n.Tr("Profile commands about libraries."),
		Example: "" +
			"  " + os.Args[0] + " profile lib add AudioZero -m my_profile\n" +
			"  " + os.Args[0] + " profile lib remove Arduino_JSON --profile my_profile\n",
	}

	libCommand.AddCommand(initLibAddCommand(srv))
	libCommand.AddCommand(initLibRemoveCommand(srv))

	return libCommand
}

func initLibAddCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	var destDir string

	addCommand := &cobra.Command{
		Use:   fmt.Sprintf("add %s[@%s]...", i18n.Tr("LIBRARY"), i18n.Tr("VERSION_NUMBER")),
		Short: i18n.Tr("Adds a library to the profile."),
		Long:  i18n.Tr("Adds a library to the profile."),
		Example: "" +
			"  " + os.Args[0] + " profile lib add AudioZero -m my_profile\n" +
			"  " + os.Args[0] + " profile lib add Arduino_JSON@0.2.0 --profile my_profile\n",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runLibAddCommand(cmd.Context(), args, srv, destDir)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return arguments.GetInstallableLibs(cmd.Context(), srv), cobra.ShellCompDirectiveDefault
		},
	}

	addCommand.Flags().StringVar(&destDir, "dest-dir", "", i18n.Tr("Location of the project file."))
	profileArg.AddToCommand(addCommand, srv)

	return addCommand
}

func runLibAddCommand(ctx context.Context, args []string, srv rpc.ArduinoCoreServiceServer, destDir string) {
	sketchPath := arguments.InitSketchPath(destDir)

	instance := instance.CreateAndInit(ctx, srv)
	libRefs, err := lib.ParseLibraryReferenceArgsAndAdjustCase(ctx, srv, instance, args)
	if err != nil {
		feedback.Fatal(i18n.Tr("Arguments error: %v", err), feedback.ErrBadArgument)
	}
	for _, lib := range libRefs {
		resp, err := srv.ProfileLibAdd(ctx, &rpc.ProfileLibAddRequest{
			Instance:    instance,
			SketchPath:  sketchPath.String(),
			ProfileName: profileArg.Get(),
			LibName:     lib.Name,
			LibVersion:  lib.Version,
		})
		if err != nil {
			feedback.Fatal(i18n.Tr("Error adding %s to the profile %s: %v", lib.Name, profileArg.Get(), err), feedback.ErrGeneric)
		}
		feedback.PrintResult(libAddResult{LibName: resp.GetLibName(), LibVersion: resp.GetLibVersion(), ProfileName: resp.ProfileName})
	}
}

func initLibRemoveCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	var destDir string

	removeCommand := &cobra.Command{
		Use:   fmt.Sprintf("remove %s[@%s]...", i18n.Tr("LIBRARY"), i18n.Tr("VERSION_NUMBER")),
		Short: i18n.Tr("Removes a library from the profile."),
		Long:  i18n.Tr("Removes a library from the profile."),
		Example: "" +
			"  " + os.Args[0] + " profile lib remove AudioZero -m my_profile\n" +
			"  " + os.Args[0] + " profile lib remove Arduino_JSON@0.2.0 --profile my_profile\n",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runLibRemoveCommand(cmd.Context(), args, srv, destDir)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return arguments.GetInstallableLibs(cmd.Context(), srv), cobra.ShellCompDirectiveDefault
		},
	}

	removeCommand.Flags().StringVar(&destDir, "dest-dir", "", i18n.Tr("Location of the project file."))
	profileArg.AddToCommand(removeCommand, srv)

	return removeCommand
}

func runLibRemoveCommand(ctx context.Context, args []string, srv rpc.ArduinoCoreServiceServer, destDir string) {
	sketchPath := arguments.InitSketchPath(destDir)

	instance := instance.CreateAndInit(ctx, srv)
	libRefs, err := lib.ParseLibraryReferenceArgsAndAdjustCase(ctx, srv, instance, args)
	if err != nil {
		feedback.Fatal(i18n.Tr("Arguments error: %v", err), feedback.ErrBadArgument)
	}
	for _, lib := range libRefs {
		resp, err := srv.ProfileLibRemove(ctx, &rpc.ProfileLibRemoveRequest{
			SketchPath:  sketchPath.String(),
			ProfileName: profileArg.Get(),
			LibName:     lib.Name,
		})
		if err != nil {
			feedback.Fatal(i18n.Tr("Error removing %s from the profile %s: %v", lib.Name, profileArg.Get(), err), feedback.ErrGeneric)
		}
		feedback.PrintResult(libRemoveResult{LibName: resp.GetLibName(), LibVersion: resp.GetLibVersion(), ProfileName: resp.ProfileName})
	}
}

type libAddResult struct {
	LibName     string `json:"library_name"`
	LibVersion  string `json:"library_version"`
	ProfileName string `json:"profile_name"`
}

func (lr libAddResult) Data() interface{} {
	return lr
}

func (lr libAddResult) String() string {
	return i18n.Tr("Profile %s: %s@%s added successfully", lr.ProfileName, lr.LibName, lr.LibVersion)
}

type libRemoveResult struct {
	LibName     string `json:"library_name"`
	LibVersion  string `json:"library_version"`
	ProfileName string `json:"profile_name"`
}

func (lr libRemoveResult) Data() interface{} {
	return lr
}

func (lr libRemoveResult) String() string {
	return i18n.Tr("Profile %s: %s@%s removed successfully", lr.ProfileName, lr.LibName, lr.LibVersion)
}
