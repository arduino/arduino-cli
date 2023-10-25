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
	"sort"
	"strings"

	"github.com/arduino/arduino-cli/commands/lib"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/feedback/result"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/table"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initListCommand() *cobra.Command {
	var all bool
	var updatable bool
	listCommand := &cobra.Command{
		Use:   fmt.Sprintf("list [%s]", tr("LIBNAME")),
		Short: tr("Shows a list of installed libraries."),
		Long: tr(`Shows a list of installed libraries.

If the LIBNAME parameter is specified the listing is limited to that specific
library. By default the libraries provided as built-in by platforms/core are
not listed, they can be listed by adding the --all flag.`),
		Example: "  " + os.Args[0] + " lib list",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			instance := instance.CreateAndInit()
			logrus.Info("Executing `arduino-cli lib list`")
			List(instance, args, all, updatable)
		},
	}
	listCommand.Flags().BoolVar(&all, "all", false, tr("Include built-in libraries (from platforms and IDE) in listing."))
	fqbn.AddToCommand(listCommand)
	listCommand.Flags().BoolVar(&updatable, "updatable", false, tr("List updatable libraries."))
	return listCommand
}

// List gets and prints a list of installed libraries.
func List(instance *rpc.Instance, args []string, all bool, updatable bool) {
	installedLibs := GetList(instance, args, all, updatable)

	installedLibsResult := make([]*result.InstalledLibrary, len(installedLibs))
	for i, v := range installedLibs {
		installedLibsResult[i] = result.NewInstalledLibrary(v)
	}
	feedback.PrintResult(installedResult{
		onlyUpdates:   updatable,
		installedLibs: installedLibsResult,
	})
	logrus.Info("Done")
}

// GetList returns a list of installed libraries.
func GetList(
	instance *rpc.Instance,
	args []string,
	all bool,
	updatable bool,
) []*rpc.InstalledLibrary {
	name := ""
	if len(args) > 0 {
		name = args[0]
	}

	res, err := lib.LibraryList(context.Background(), &rpc.LibraryListRequest{
		Instance:  instance,
		All:       all,
		Updatable: updatable,
		Name:      name,
		Fqbn:      fqbn.String(),
	})
	if err != nil {
		feedback.Fatal(tr("Error listing libraries: %v", err), feedback.ErrGeneric)
	}

	libs := []*rpc.InstalledLibrary{}
	if fqbn.String() == "" {
		libs = res.GetInstalledLibraries()
	} else {
		for _, lib := range res.GetInstalledLibraries() {
			if lib.Library.CompatibleWith[fqbn.String()] {
				libs = append(libs, lib)
			}
		}
	}

	// To uniform the output to other commands, when there are no result
	// print out an empty slice.
	if libs == nil {
		libs = []*rpc.InstalledLibrary{}
	}

	return libs
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type installedResult struct {
	onlyUpdates   bool
	installedLibs []*result.InstalledLibrary
}

func (ir installedResult) Data() interface{} {
	return ir.installedLibs
}

func (ir installedResult) String() string {
	if len(ir.installedLibs) == 0 {
		if ir.onlyUpdates {
			return tr("No libraries update is available.")
		}
		return tr("No libraries installed.")
	}
	sort.Slice(ir.installedLibs, func(i, j int) bool {
		return strings.ToLower(ir.installedLibs[i].Library.Name) < strings.ToLower(ir.installedLibs[j].Library.Name) ||
			strings.ToLower(ir.installedLibs[i].Library.ContainerPlatform) < strings.ToLower(ir.installedLibs[j].Library.ContainerPlatform)
	})

	t := table.New()
	t.SetHeader(tr("Name"), tr("Installed"), tr("Available"), tr("Location"), tr("Description"))
	t.SetColumnWidthMode(1, table.Average)
	t.SetColumnWidthMode(2, table.Average)
	t.SetColumnWidthMode(4, table.Average)

	lastName := ""
	for _, libMeta := range ir.installedLibs {
		if libMeta == nil {
			continue
		}
		lib := libMeta.Library
		name := lib.Name
		if name == lastName {
			name = ` "`
		} else {
			lastName = name
		}

		location := string(lib.Location)
		if lib.ContainerPlatform != "" {
			location = lib.ContainerPlatform
		}

		available := ""
		sentence := ""
		if libMeta.Release != nil {
			available = libMeta.Release.Version
			sentence = lib.Sentence
		}

		if available == "" {
			available = "-"
		}
		if sentence == "" {
			sentence = "-"
		} else if len(sentence) > 40 {
			sentence = sentence[:37] + "..."
		}
		t.AddRow(name, lib.Version, available, location, sentence)
	}

	return t.Render()
}
