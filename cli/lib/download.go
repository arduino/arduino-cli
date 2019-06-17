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
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/spf13/cobra"
)

func initDownloadCommand() *cobra.Command {
	downloadCommand := &cobra.Command{
		Use:   "download [LIBRARY_NAME(S)]",
		Short: "Downloads one or more libraries without installing them.",
		Long:  "Downloads one or more libraries without installing them.",
		Example: "" +
			"  " + cli.VersionInfo.Application + " lib download AudioZero       # for the latest version.\n" +
			"  " + cli.VersionInfo.Application + " lib download AudioZero@1.0.0 # for a specific version.",
		Args: cobra.MinimumNArgs(1),
		Run:  runDownloadCommand,
	}
	return downloadCommand
}

func runDownloadCommand(cmd *cobra.Command, args []string) {
	instance := cli.CreateInstaceIgnorePlatformIndexErrors()
	pairs, err := librariesindex.ParseArgs(args)
	if err != nil {
		formatter.PrintError(err, "Arguments error")
		os.Exit(cli.ErrBadArgument)
	}
	for _, library := range pairs {
		libraryDownloadReq := &rpc.LibraryDownloadReq{
			Instance: instance,
			Name:     library.Name,
			Version:  library.Version.String(),
		}
		_, err := lib.LibraryDownload(context.Background(), libraryDownloadReq, cli.OutputProgressBar(),
			cli.HTTPClientHeader)
		if err != nil {
			formatter.PrintError(err, "Error downloading "+library.String())
			os.Exit(cli.ErrNetwork)
		}
	}
}
