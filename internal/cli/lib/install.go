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
	"strings"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/version"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	semver "go.bug.st/relaxed-semver"
)

func initInstallCommand(srv rpc.ArduinoCoreServiceServer, settings *rpc.Configuration) *cobra.Command {
	var noDeps bool
	var noOverwrite bool
	var gitURL bool
	var zipPath bool
	var useBuiltinLibrariesDir bool
	enableUnsafeInstall := settings.GetLibrary().GetEnableUnsafeInstall()
	installCommand := &cobra.Command{
		Use:   fmt.Sprintf("install %s[@%s]...", i18n.Tr("LIBRARY"), i18n.Tr("VERSION_NUMBER")),
		Short: i18n.Tr("Installs one or more specified libraries into the system."),
		Long:  i18n.Tr("Installs one or more specified libraries into the system."),
		Example: "" +
			"  " + os.Args[0] + " lib install AudioZero       # " + i18n.Tr("for the latest version.") + "\n" +
			"  " + os.Args[0] + " lib install AudioZero@1.0.0 # " + i18n.Tr("for the specific version.") + "\n" +
			"  " + os.Args[0] + " lib install --git-url https://github.com/arduino-libraries/WiFi101.git https://github.com/arduino-libraries/ArduinoBLE.git\n" +
			"  " + os.Args[0] + " lib install --git-url https://github.com/arduino-libraries/WiFi101.git#0.16.0 # " + i18n.Tr("for the specific version.") + "\n" +
			"  " + os.Args[0] + " lib install --zip-path /path/to/WiFi101.zip /path/to/ArduinoBLE.zip\n",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runInstallCommand(cmd.Context(), srv, args, noDeps, noOverwrite, gitURL, zipPath, useBuiltinLibrariesDir, enableUnsafeInstall)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return arguments.GetInstallableLibs(cmd.Context(), srv), cobra.ShellCompDirectiveDefault
		},
	}
	installCommand.Flags().BoolVar(&noDeps, "no-deps", false, i18n.Tr("Do not install dependencies."))
	installCommand.Flags().BoolVar(&noOverwrite, "no-overwrite", false, i18n.Tr("Do not overwrite already installed libraries."))
	installCommand.Flags().BoolVar(&gitURL, "git-url", false, i18n.Tr("Enter git url for libraries hosted on repositories"))
	installCommand.Flags().BoolVar(&zipPath, "zip-path", false, i18n.Tr("Enter a path to zip file"))
	installCommand.Flags().BoolVar(&useBuiltinLibrariesDir, "install-in-builtin-dir", false, i18n.Tr("Install libraries in the IDE-Builtin directory"))
	return installCommand
}

func runInstallCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer, args []string, noDeps bool, noOverwrite bool, gitURL bool, zipPath bool, useBuiltinLibrariesDir bool, enableUnsafeInstall bool) {
	instance := instance.CreateAndInit(ctx, srv)
	logrus.Info("Executing `arduino-cli lib install`")

	if zipPath || gitURL {
		if !enableUnsafeInstall {
			documentationURL := "https://arduino.github.io/arduino-cli/latest/configuration/#configuration-keys"
			_, err := semver.Parse(version.VersionInfo.VersionString)
			if err == nil {
				split := strings.Split(version.VersionInfo.VersionString, ".")
				documentationURL = fmt.Sprintf("https://arduino.github.io/arduino-cli/%s.%s/configuration/#configuration-keys", split[0], split[1])
			}
			feedback.Fatal(i18n.Tr("--git-url and --zip-path are disabled by default, for more information see: %v", documentationURL), feedback.ErrGeneric)
		}
		feedback.Print(i18n.Tr("--git-url and --zip-path flags allow installing untrusted files, use it at your own risk."))

		if useBuiltinLibrariesDir {
			feedback.Fatal(i18n.Tr("--git-url or --zip-path can't be used with --install-in-builtin-dir"), feedback.ErrGeneric)
		}
	}

	if zipPath {
		for _, path := range args {
			req := &rpc.ZipLibraryInstallRequest{
				Instance:  instance,
				Path:      path,
				Overwrite: !noOverwrite,
			}
			stream := commands.ZipLibraryInstallStreamResponseToCallbackFunction(ctx, feedback.TaskProgress())
			if err := srv.ZipLibraryInstall(req, stream); err != nil {
				feedback.Fatal(i18n.Tr("Error installing Zip Library: %v", err), feedback.ErrGeneric)
			}
		}
		return
	}

	if gitURL {
		for _, url := range args {
			if url == "." {
				wd, err := paths.Getwd()
				if err != nil {
					feedback.Fatal(i18n.Tr("Couldn't get current working directory: %v", err), feedback.ErrGeneric)
				}
				url = wd.String()
			}
			req := &rpc.GitLibraryInstallRequest{
				Instance:  instance,
				Url:       url,
				Overwrite: !noOverwrite,
			}
			stream := commands.GitLibraryInstallStreamResponseToCallbackFunction(ctx, feedback.TaskProgress())
			if err := srv.GitLibraryInstall(req, stream); err != nil {
				feedback.Fatal(i18n.Tr("Error installing Git Library: %v", err), feedback.ErrGeneric)
			}
		}
		return
	}

	libRefs, err := ParseLibraryReferenceArgsAndAdjustCase(ctx, srv, instance, args)
	if err != nil {
		feedback.Fatal(i18n.Tr("Arguments error: %v", err), feedback.ErrBadArgument)
	}

	for _, libRef := range libRefs {
		installLocation := rpc.LibraryInstallLocation_LIBRARY_INSTALL_LOCATION_USER
		if useBuiltinLibrariesDir {
			installLocation = rpc.LibraryInstallLocation_LIBRARY_INSTALL_LOCATION_BUILTIN
		}
		libraryInstallRequest := &rpc.LibraryInstallRequest{
			Instance:        instance,
			Name:            libRef.Name,
			Version:         libRef.Version,
			NoDeps:          noDeps,
			NoOverwrite:     noOverwrite,
			InstallLocation: installLocation,
		}
		stream := commands.LibraryInstallStreamResponseToCallbackFunction(ctx, feedback.ProgressBar(), feedback.TaskProgress())
		if err := srv.LibraryInstall(libraryInstallRequest, stream); err != nil {
			feedback.Fatal(i18n.Tr("Error installing %s: %v", libRef.Name, err), feedback.ErrGeneric)
		}
	}
}
