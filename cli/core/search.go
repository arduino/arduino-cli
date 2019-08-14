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

package core

import (
	"context"
	"os"
	"sort"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/core"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/arduino/arduino-cli/table"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initSearchCommand() *cobra.Command {
	searchCommand := &cobra.Command{
		Use:     "search <keywords...>",
		Short:   "Search for a core in the package index.",
		Long:    "Search for a core in the package index using the specified keywords.",
		Example: "  " + os.Args[0] + " core search MKRZero -v",
		Args:    cobra.MinimumNArgs(1),
		Run:     runSearchCommand,
	}
	return searchCommand
}

func runSearchCommand(cmd *cobra.Command, args []string) {
	instance := instance.CreateInstance()
	logrus.Info("Executing `arduino core search`")

	arguments := strings.ToLower(strings.Join(args, " "))
	feedback.Printf("Searching for platforms matching '%s'", arguments)

	resp, err := core.PlatformSearch(context.Background(), &rpc.PlatformSearchReq{
		Instance:   instance,
		SearchArgs: arguments,
	})
	if err != nil {
		feedback.Errorf("Error saerching for platforms: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	coreslist := resp.GetSearchOutput()
	if globals.OutputFormat == "json" {
		feedback.PrintJSON(coreslist)
	} else {
		outputSearchCores(coreslist)
	}
}

func outputSearchCores(cores []*rpc.Platform) {
	if len(cores) > 0 {
		t := table.New()
		t.SetHeader("ID", "Version", "Name")
		sort.Slice(cores, func(i, j int) bool {
			return cores[i].ID < cores[j].ID
		})
		for _, item := range cores {
			t.AddRow(item.GetID(), item.GetLatest(), item.GetName())
		}
		feedback.Print(t.Render())
	} else {
		feedback.Print("No platforms matching your search.")
	}
}
