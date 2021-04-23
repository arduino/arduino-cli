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

package core

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/table"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	allVersions bool
)

func initSearchCommand() *cobra.Command {
	searchCommand := &cobra.Command{
		Use:     "search <keywords...>",
		Short:   "Search for a core in Boards Manager.",
		Long:    "Search for a core in Boards Manager using the specified keywords.",
		Example: "  " + os.Args[0] + " core search MKRZero -a -v",
		Args:    cobra.ArbitraryArgs,
		Run:     runSearchCommand,
	}
	searchCommand.Flags().BoolVarP(&allVersions, "all", "a", false, "Show all available core versions.")

	return searchCommand
}

// indexUpdateInterval specifies the time threshold over which indexes are updated
const indexUpdateInterval = "24h"

func runSearchCommand(cmd *cobra.Command, args []string) {
	inst := instance.CreateAndInit()

	if indexesNeedUpdating(indexUpdateInterval) {
		_, err := commands.UpdateIndex(context.Background(), &rpc.UpdateIndexRequest{
			Instance: inst,
		}, output.ProgressBar())
		if err != nil {
			feedback.Errorf("Error updating index: %v", err)
			os.Exit(errorcodes.ErrGeneric)
		}
	}

	arguments := strings.ToLower(strings.Join(args, " "))
	logrus.Infof("Executing `arduino core search` with args: '%s'", arguments)

	resp, err := core.PlatformSearch(&rpc.PlatformSearchRequest{
		Instance:    inst,
		SearchArgs:  arguments,
		AllVersions: allVersions,
	})
	if err != nil {
		feedback.Errorf("Error searching for platforms: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	coreslist := resp.GetSearchOutput()
	feedback.PrintResult(searchResults{coreslist})
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type searchResults struct {
	platforms []*rpc.Platform
}

func (sr searchResults) Data() interface{} {
	return sr.platforms
}

func (sr searchResults) String() string {
	if len(sr.platforms) > 0 {
		t := table.New()
		t.SetHeader("ID", "Version", "Name")
		for _, item := range sr.platforms {
			name := item.GetName()
			if item.Deprecated {
				name = fmt.Sprintf("[DEPRECATED] %s", name)
			}
			t.AddRow(item.GetId(), item.GetLatest(), name)
		}
		return t.Render()
	}
	return "No platforms matching your search."
}

// indexesNeedUpdating returns whether one or more index files need updating.
// A duration string must be provided to calculate the time threshold
// used to update the indexes, if the duration is not valid a default
// of 24 hours is used.
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
func indexesNeedUpdating(duration string) bool {
	indexpath := paths.New(configuration.Settings.GetString("directories.Data"))

	now := time.Now()
	modTimeThreshold, err := time.ParseDuration(duration)
	// Not the most elegant way of handling this error
	// but it does its job
	if err != nil {
		modTimeThreshold, _ = time.ParseDuration("24h")
	}

	urls := []string{globals.DefaultIndexURL}
	urls = append(urls, configuration.Settings.GetStringSlice("board_manager.additional_urls")...)
	for _, u := range urls {
		URL, err := utils.URLParse(u)
		if err != nil {
			continue
		}

		if URL.Scheme == "file" {
			// No need to update local files
			continue
		}

		coreIndexPath := indexpath.Join(path.Base(URL.Path))
		if coreIndexPath.NotExist() {
			return true
		}

		info, err := coreIndexPath.Stat()
		if err != nil {
			return true
		}

		if now.After(info.ModTime().Add(modTimeThreshold)) {
			return true
		}
	}
	return false
}
