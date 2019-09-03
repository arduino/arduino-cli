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
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/lib"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/arduino/arduino-cli/table"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

func initListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:     "list",
		Short:   "Shows a list of all installed libraries.",
		Long:    "Shows a list of all installed libraries.",
		Example: "  " + os.Args[0] + " lib list",
		Args:    cobra.NoArgs,
		Run:     runListCommand,
	}
	listCommand.Flags().BoolVar(&listFlags.all, "all", false, "Include built-in libraries (from platforms and IDE) in listing.")
	listCommand.Flags().BoolVar(&listFlags.updatable, "updatable", false, "List updatable libraries.")
	return listCommand
}

var listFlags struct {
	all       bool
	updatable bool
}

func runListCommand(cmd *cobra.Command, args []string) {
	instance := instance.CreateInstaceIgnorePlatformIndexErrors()
	logrus.Info("Listing")

	res, err := lib.LibraryList(context.Background(), &rpc.LibraryListReq{
		Instance:  instance,
		All:       listFlags.all,
		Updatable: listFlags.updatable,
	})
	if err != nil {
		feedback.Errorf("Error listing Libraries: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	libs := res.GetInstalledLibrary()
	feedback.PrintResult(installedResult{libs})

	logrus.Info("Done")
}

// ouput from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type installedResult struct {
	installedLibs []*rpc.InstalledLibrary
}

func (ir installedResult) Data() interface{} {
	return ir.installedLibs
}

func (ir installedResult) String() string {
	if ir.installedLibs == nil || len(ir.installedLibs) == 0 {
		return ""
	}

	t := table.New()
	t.SetHeader("Name", "Installed", "Available", "Location")

	lastName := ""
	for _, libMeta := range ir.installedLibs {
		lib := libMeta.GetLibrary()
		name := lib.Name
		if name == lastName {
			name = ` "`
		} else {
			lastName = name
		}

		location := lib.GetLocation()
		if lib.ContainerPlatform != "" {
			location = lib.GetContainerPlatform()
		}

		if libMeta.GetRelease() != nil {
			available := libMeta.GetRelease().GetVersion()
			if available != "" {
				t.AddRow(name, lib.Version, available, location)
			} else {
				t.AddRow(name, lib.Version, "-", location)
			}
		}
	}

	return t.Render()
}
