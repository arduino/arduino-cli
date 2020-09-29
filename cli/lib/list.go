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
	"os"
	"sort"
	"strings"

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
		Use:   "list [LIBNAME]",
		Short: "Shows a list of installed libraries.",
		Long: "Shows a list of installed libraries.\n\n" +
			"If the LIBNAME parameter is specified the listing is limited to that specific\n" +
			"library. By default the libraries provided as built-in by platforms/core are\n" +
			"not listed, they can be listed by adding the --all flag.",
		Example: "  " + os.Args[0] + " lib list",
		Args:    cobra.MaximumNArgs(1),
		Run:     runListCommand,
	}
	listCommand.Flags().BoolVar(&listFlags.all, "all", false, "Include built-in libraries (from platforms and IDE) in listing.")
	listCommand.Flags().StringVarP(&listFlags.fqbn, "fqbn", "b", "", "Show libraries for the specified board FQBN.")
	listCommand.Flags().BoolVar(&listFlags.updatable, "updatable", false, "List updatable libraries.")
	return listCommand
}

var listFlags struct {
	all       bool
	updatable bool
	fqbn      string
}

func runListCommand(cmd *cobra.Command, args []string) {
	instance := instance.CreateInstanceIgnorePlatformIndexErrors()
	logrus.Info("Listing")

	name := ""
	if len(args) > 0 {
		name = args[0]
	}

	res, err := lib.LibraryList(context.Background(), &rpc.LibraryListReq{
		Instance:  instance,
		All:       listFlags.all,
		Updatable: listFlags.updatable,
		Name:      name,
		Fqbn:      listFlags.fqbn,
	})
	if err != nil {
		feedback.Errorf("Error listing Libraries: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	libs := []*rpc.InstalledLibrary{}
	if listFlags.fqbn == "" {
		libs = res.GetInstalledLibrary()
	} else {
		for _, lib := range res.GetInstalledLibrary() {
			if lib.Library.CompatibleWith[listFlags.fqbn] {
				libs = append(libs, lib)
			}
		}
	}

	// To uniform the output to other commands, when there are no result
	// print out an empty slice.
	if libs == nil {
		libs = []*rpc.InstalledLibrary{}
	}

	feedback.PrintResult(installedResult{libs})
	logrus.Info("Done")
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type installedResult struct {
	installedLibs []*rpc.InstalledLibrary
}

func (ir installedResult) Data() interface{} {
	return ir.installedLibs
}

func (ir installedResult) String() string {
	if ir.installedLibs == nil || len(ir.installedLibs) == 0 {
		return "No libraries installed."
	}
	sort.Slice(ir.installedLibs, func(i, j int) bool {
		return strings.ToLower(ir.installedLibs[i].Library.Name) < strings.ToLower(ir.installedLibs[j].Library.Name) ||
			strings.ToLower(ir.installedLibs[i].Library.ContainerPlatform) < strings.ToLower(ir.installedLibs[j].Library.ContainerPlatform)
	})

	t := table.New()
	t.SetHeader("Name", "Installed", "Available", "Location", "Description")
	t.SetColumnWidthMode(1, table.Average)
	t.SetColumnWidthMode(2, table.Average)
	t.SetColumnWidthMode(4, table.Average)

	lastName := ""
	for _, libMeta := range ir.installedLibs {
		lib := libMeta.GetLibrary()
		name := lib.Name
		if name == lastName {
			name = ` "`
		} else {
			lastName = name
		}

		location := lib.GetLocation().String()
		if lib.ContainerPlatform != "" {
			location = lib.GetContainerPlatform()
		}

		if libMeta.GetRelease() != nil {
			available := libMeta.GetRelease().GetVersion()
			if available == "" {
				available = "-"
			}
			sentence := lib.Sentence
			if sentence == "" {
				sentence = "-"
			} else if len(sentence) > 40 {
				sentence = sentence[:37] + "..."
			}
			t.AddRow(name, lib.Version, available, location, sentence)
		}
	}

	return t.Render()
}
