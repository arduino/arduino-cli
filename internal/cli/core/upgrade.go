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

package core

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initUpgradeCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	var postInstallFlags arguments.PrePostScriptsFlags
	upgradeCommand := &cobra.Command{
		Use:   fmt.Sprintf("upgrade [%s:%s] ...", i18n.Tr("PACKAGER"), i18n.Tr("ARCH")),
		Short: i18n.Tr("Upgrades one or all installed platforms to the latest version."),
		Long:  i18n.Tr("Upgrades one or all installed platforms to the latest version."),
		Example: "" +
			"  # " + i18n.Tr("upgrade everything to the latest version") + "\n" +
			"  " + os.Args[0] + " core upgrade\n\n" +
			"  # " + i18n.Tr("upgrade arduino:samd to the latest version") + "\n" +
			"  " + os.Args[0] + " core upgrade arduino:samd",
		Run: func(cmd *cobra.Command, args []string) {
			runUpgradeCommand(cmd.Context(), srv, args, postInstallFlags.DetectSkipPostInstallValue(), postInstallFlags.DetectSkipPreUninstallValue())
		},
	}
	postInstallFlags.AddToCommand(upgradeCommand)
	return upgradeCommand
}

func runUpgradeCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer, args []string, skipPostInstall bool, skipPreUninstall bool) {
	logrus.Info("Executing `arduino-cli core upgrade`")
	inst := instance.CreateAndInit(ctx, srv)
	Upgrade(ctx, srv, inst, args, skipPostInstall, skipPreUninstall)
}

// Upgrade upgrades one or all installed platforms to the latest version.
func Upgrade(ctx context.Context, srv rpc.ArduinoCoreServiceServer, inst *rpc.Instance, args []string, skipPostInstall bool, skipPreUninstall bool) {
	// if no platform was passed, upgrade allthethings
	if len(args) == 0 {
		platforms, err := srv.PlatformSearch(ctx, &rpc.PlatformSearchRequest{
			Instance: inst,
		})
		if err != nil {
			feedback.Fatal(i18n.Tr("Error retrieving core list: %v", err), feedback.ErrGeneric)
		}

		targets := []*rpc.Platform{}
		for _, platform := range platforms.GetSearchOutput() {
			if platform.GetInstalledVersion() == "" {
				continue
			}
			// if it's not updatable, skip it
			latestRelease := platform.GetLatestRelease()
			if latestRelease != nil && platform.GetInstalledVersion() != latestRelease.GetVersion() {
				targets = append(targets, &rpc.Platform{
					Metadata: platform.GetMetadata(),
					Release:  latestRelease,
				})
			}
		}

		if len(targets) == 0 {
			feedback.Print(i18n.Tr("All the cores are already at the latest version"))
			return
		}

		for _, t := range targets {
			args = append(args, t.GetMetadata().GetId())
		}
	}

	warningMissingIndex := func(platform *rpc.Platform) {
		if platform == nil {
			return
		}
		if !platform.GetMetadata().GetIndexed() {
			feedback.Warning(i18n.Tr("missing package index for %s, future updates cannot be guaranteed", platform.GetMetadata().GetId()))
		}
	}

	// proceed upgrading, if anything is upgradable
	platformsRefs, err := arguments.ParseReferences(ctx, srv, args)
	if err != nil {
		feedback.Fatal(i18n.Tr("Invalid argument passed: %v", err), feedback.ErrBadArgument)
	}

	hasBadArguments := false
	for i, platformRef := range platformsRefs {
		if platformRef.Version != "" {
			feedback.Warning(i18n.Tr("Invalid item %s", args[i]))
			hasBadArguments = true
			continue
		}

		r := &rpc.PlatformUpgradeRequest{
			Instance:         inst,
			PlatformPackage:  platformRef.PackageName,
			Architecture:     platformRef.Architecture,
			SkipPostInstall:  skipPostInstall,
			SkipPreUninstall: skipPreUninstall,
		}
		stream, respCB := commands.PlatformUpgradeStreamResponseToCallbackFunction(ctx, feedback.ProgressBar(), feedback.TaskProgress())
		err := srv.PlatformUpgrade(r, stream)
		warningMissingIndex(respCB().GetPlatform())
		if err != nil {
			var alreadyAtLatestVersionErr *cmderrors.PlatformAlreadyAtTheLatestVersionError
			if errors.As(err, &alreadyAtLatestVersionErr) {
				feedback.Warning(err.Error())
				continue
			}

			feedback.Fatal(i18n.Tr("Error during upgrade: %v", err), feedback.ErrGeneric)
		}
	}

	if hasBadArguments {
		feedback.Fatal(i18n.Tr("Some upgrades failed, please check the output for details."), feedback.ErrBadArgument)
	}

	feedback.PrintResult(&platformUpgradeResult{})
}

// This is needed so we can print warning messages in case users use --json
type platformUpgradeResult struct{}

// Data implements feedback.Result.
func (r *platformUpgradeResult) Data() interface{} {
	return r
}

// String implements feedback.Result.
func (r *platformUpgradeResult) String() string {
	return ""
}
