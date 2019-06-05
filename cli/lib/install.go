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

	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/cli"
	"github.com/arduino/arduino-cli/commands/lib"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/global"
	"github.com/arduino/arduino-cli/rpc"
	"github.com/spf13/cobra"
)

func initInstallCommand() *cobra.Command {
	installCommand := &cobra.Command{
		Use:   "install LIBRARY[@VERSION_NUMBER](S)",
		Short: "Installs one of more specified libraries into the system.",
		Long:  "Installs one or more specified libraries into the system.",
		Example: "" +
			"  " + global.GetAppName() + " lib install AudioZero       # for the latest version.\n" +
			"  " + global.GetAppName() + " lib install AudioZero@1.0.0 # for the specific version.",
		Args: cobra.MinimumNArgs(1),
		Run:  runInstallCommand,
	}
	return installCommand
}

func runInstallCommand(cmd *cobra.Command, args []string) {
	instance := cli.CreateInstaceIgnorePlatformIndexErrors()
	refs, err := librariesindex.ParseArgs(args)
	if err != nil {
		formatter.PrintError(err, "Arguments error")
		os.Exit(cli.ErrBadArgument)
	}
	for _, library := range refs {
		err := lib.LibraryInstall(context.Background(), &rpc.LibraryInstallReq{
			Instance: instance,
			Name:     library.Name,
			Version:  library.Version.String(),
		}, cli.OutputProgressBar(), cli.OutputTaskProgress())
		if err != nil {
			formatter.PrintError(err, "Error installing "+library.String())
			os.Exit(cli.ErrGeneric)
		}
	}
}
