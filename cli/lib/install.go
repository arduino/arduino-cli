/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package lib

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands/lib"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/spf13/cobra"
)

func initInstallCommand() *cobra.Command {
	installCommand := &cobra.Command{
		Use:   "install LIBRARY[@VERSION_NUMBER](S)",
		Short: "Installs one of more specified libraries into the system.",
		Long:  "Installs one or more specified libraries into the system.",
		Example: "" +
			"  " + os.Args[0] + " lib install AudioZero       # for the latest version.\n" +
			"  " + os.Args[0] + " lib install AudioZero@1.0.0 # for the specific version.",
		Args: cobra.MinimumNArgs(1),
		Run:  runInstallCommand,
	}
	return installCommand
}

func runInstallCommand(cmd *cobra.Command, args []string) {
	instance := instance.CreateInstaceIgnorePlatformIndexErrors()
	refs, err := globals.ParseReferenceArgs(args, false)
	if err != nil {
		feedback.Errorf("Arguments error: %v", err)
		os.Exit(errorcodes.ErrBadArgument)
	}

	for _, library := range refs {
		libraryInstallReq := &rpc.LibraryInstallReq{
			Instance: instance,
			Name:     library.PackageName,
			Version:  library.Version,
		}
		err := lib.LibraryInstall(context.Background(), libraryInstallReq, output.ProgressBar(),
			output.TaskProgress(), globals.HTTPClientHeader)
		if err != nil {
			feedback.Errorf("Error installing %s: %v", library, err)
			os.Exit(errorcodes.ErrGeneric)
		}
	}
}
