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

	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/feedback/result"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	fqbn arguments.Fqbn
)

func initExamplesCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	examplesCommand := &cobra.Command{
		Use:     fmt.Sprintf("examples [%s]", i18n.Tr("LIBRARY_NAME")),
		Short:   i18n.Tr("Shows the list of the examples for libraries."),
		Long:    i18n.Tr("Shows the list of the examples for libraries. A name may be given as argument to search a specific library."),
		Example: "  " + os.Args[0] + " lib examples Wire",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runExamplesCommand(cmd.Context(), srv, args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return arguments.GetInstalledLibraries(cmd.Context(), srv), cobra.ShellCompDirectiveDefault
		},
	}
	fqbn.AddToCommand(examplesCommand, srv)
	return examplesCommand
}

func runExamplesCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer, args []string) {
	logrus.Info("Executing `arduino-cli lib examples`")
	instance := instance.CreateAndInit(ctx, srv)

	name := ""
	if len(args) > 0 {
		name = args[0]
	}

	res, err := srv.LibraryList(ctx, &rpc.LibraryListRequest{
		Instance: instance,
		All:      true,
		Name:     name,
		Fqbn:     fqbn.String(),
	})
	if err != nil {
		feedback.Fatal(i18n.Tr("Error getting libraries info: %v", err), feedback.ErrGeneric)
	}

	found := []*libraryExamples{}
	for _, lib := range res.GetInstalledLibraries() {
		found = append(found, &libraryExamples{
			Library:  result.NewLibrary(lib.GetLibrary()),
			Examples: lib.GetLibrary().GetExamples(),
		})
	}

	feedback.PrintResult(libraryExamplesResult{found})
	logrus.Info("Done")
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation

type libraryExamples struct {
	Library  *result.Library `json:"library"`
	Examples []string        `json:"examples"`
}

type libraryExamplesResult struct {
	Examples []*libraryExamples `json:"examples"`
}

func (ir libraryExamplesResult) Data() interface{} {
	return ir
}

func (ir libraryExamplesResult) String() string {
	if ir.Examples == nil || len(ir.Examples) == 0 {
		return i18n.Tr("No libraries found.")
	}

	sort.Slice(ir.Examples, func(i, j int) bool {
		return strings.ToLower(ir.Examples[i].Library.Name) < strings.ToLower(ir.Examples[j].Library.Name)
	})

	res := []string{}
	for _, lib := range ir.Examples {
		name := lib.Library.Name
		if lib.Library.ContainerPlatform != "" {
			name += " (" + lib.Library.ContainerPlatform + ")"
		} else if lib.Library.Location != result.LibraryLocationUser {
			name += " (" + string(lib.Library.Location) + ")"
		}
		r := i18n.Tr("Examples for library %s", color.GreenString("%s", name)) + "\n"
		sort.Slice(lib.Examples, func(i, j int) bool {
			return strings.ToLower(lib.Examples[i]) < strings.ToLower(lib.Examples[j])
		})
		for _, example := range lib.Examples {
			examplePath := paths.New(example)
			r += fmt.Sprintf("  - %s%s\n",
				color.New(color.Faint).Sprintf("%s%c", examplePath.Parent(), os.PathSeparator),
				examplePath.Base())
		}
		res = append(res, r)
	}

	return strings.Join(res, "\n")
}
