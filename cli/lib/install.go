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

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands/lib"
	"github.com/arduino/arduino-cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/spf13/cobra"
)

func initInstallCommand() *cobra.Command {
	installCommand := &cobra.Command{
		Use:   "install LIBRARY[@VERSION_NUMBER](S)",
		Short: "Installs one or more specified libraries into the system.",
		Long:  "Installs one or more specified libraries into the system.",
		Example: "" +
			"  " + os.Args[0] + " lib install AudioZero       # for the latest version.\n" +
			"  " + os.Args[0] + " lib install AudioZero@1.0.0 # for the specific version.\n" +
			"  " + os.Args[0] + " lib install --git-url https://github.com/arduino-libraries/WiFi101.git https://github.com/arduino-libraries/ArduinoBLE.git\n" +
			"  " + os.Args[0] + " lib install --zip-path /path/to/WiFi101.zip /path/to/ArduinoBLE.zip\n",
		Args: cobra.MinimumNArgs(1),
		Run:  runInstallCommand,
	}
	installCommand.Flags().BoolVar(&installFlags.noDeps, "no-deps", false, "Do not install dependencies.")
	installCommand.Flags().BoolVar(&installFlags.gitURL, "git-url", false, "Enter git url for libraries hosted on repositories")
	installCommand.Flags().BoolVar(&installFlags.zipPath, "zip-path", false, "Enter a path to zip file")
	return installCommand
}

var installFlags struct {
	noDeps  bool
	gitURL  bool
	zipPath bool
}

func runInstallCommand(cmd *cobra.Command, args []string) {
	instance := instance.CreateAndInit()

	if installFlags.zipPath || installFlags.gitURL {
		if !configuration.Settings.GetBool("library.enable_unsafe_install") {
			documentationURL := "https://arduino.github.io/arduino-cli/latest/configuration/#configuration-keys"
			if !strings.Contains(globals.VersionInfo.VersionString, "git") {
				split := strings.Split(globals.VersionInfo.VersionString, ".")
				documentationURL = fmt.Sprintf("https://arduino.github.io/arduino-cli/%s.%s/configuration/#configuration-keys", split[0], split[1])
			}
			feedback.Errorf("--git-url and --zip-path are disabled by default, for more information see: %v", documentationURL)
			os.Exit(errorcodes.ErrGeneric)
		}
		feedback.Print("--git-url and --zip-path flags allow installing untrusted files, use it at your own risk.")
	}

	if installFlags.zipPath {
		for _, path := range args {
			err := lib.ZipLibraryInstall(context.Background(), &rpc.ZipLibraryInstallRequest{
				Instance:  instance,
				Path:      path,
				Overwrite: true,
			}, output.TaskProgress())
			if err != nil {
				feedback.Errorf("Error installing Zip Library: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		}
		return
	}

	if installFlags.gitURL {
		for _, url := range args {
			if url == "." {
				wd, err := paths.Getwd()
				if err != nil {
					feedback.Errorf("Couldn't get current working directory: %v", err)
					os.Exit(errorcodes.ErrGeneric)
				}
				url = wd.String()
			}
			err := lib.GitLibraryInstall(context.Background(), &rpc.GitLibraryInstallRequest{
				Instance:  instance,
				Url:       url,
				Overwrite: true,
			}, output.TaskProgress())
			if err != nil {
				feedback.Errorf("Error installing Git Library: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		}
		return
	}

	libRefs, err := ParseLibraryReferenceArgsAndAdjustCase(instance, args)
	if err != nil {
		feedback.Errorf("Arguments error: %v", err)
		os.Exit(errorcodes.ErrBadArgument)
	}

	for _, libRef := range libRefs {
		libraryInstallRequest := &rpc.LibraryInstallRequest{
			Instance: instance,
			Name:     libRef.Name,
			Version:  libRef.Version,
			NoDeps:   installFlags.noDeps,
		}
		err := lib.LibraryInstall(context.Background(), libraryInstallRequest, output.ProgressBar(), output.TaskProgress())
		if err != nil {
			feedback.Errorf("Error installing %s: %v", libRef.Name, err)
			os.Exit(errorcodes.ErrGeneric)
		}
	}
}
