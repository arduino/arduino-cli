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

	"github.com/arduino/arduino-cli/arduino/globals"
	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/table"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	allVersions bool
)

func initSearchCommand() *cobra.Command {
	searchCommand := &cobra.Command{
		Use:     fmt.Sprintf("search <%s...>", tr("keywords")),
		Short:   tr("Search for a core in Boards Manager."),
		Long:    tr("Search for a core in Boards Manager using the specified keywords."),
		Example: "  " + os.Args[0] + " core search MKRZero -a -v",
		Args:    cobra.ArbitraryArgs,
		Run:     runSearchCommand,
	}
	searchCommand.Flags().BoolVarP(&allVersions, "all", "a", false, tr("Show all available core versions."))

	return searchCommand
}

// indexUpdateInterval specifies the time threshold over which indexes are updated
const indexUpdateInterval = "24h"

func runSearchCommand(cmd *cobra.Command, args []string) {
	inst, status := instance.Create()
	if status != nil {
		feedback.Errorf(tr("Error creating instance: %v"), status)
		os.Exit(errorcodes.ErrGeneric)
	}

	if indexesNeedUpdating(indexUpdateInterval) {
		err := commands.UpdateIndex(context.Background(), &rpc.UpdateIndexRequest{Instance: inst}, feedback.ProgressBar())
		if err != nil {
			os.Exit(errorcodes.ErrGeneric)
		}
	}

	instance.Init(inst)

	arguments := strings.ToLower(strings.Join(args, " "))
	logrus.Infof("Executing `arduino-cli core search` with args: '%s'", arguments)

	resp, err := core.PlatformSearch(&rpc.PlatformSearchRequest{
		Instance:    inst,
		SearchArgs:  arguments,
		AllVersions: allVersions,
	})
	if err != nil {
		feedback.Errorf(tr("Error searching for platforms: %v"), err)
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
		t.SetHeader(tr("ID"), tr("Version"), tr("Name"))
		for _, item := range sr.platforms {
			name := item.GetName()
			if item.Deprecated {
				name = fmt.Sprintf("[%s] %s", tr("DEPRECATED"), name)
			}
			t.AddRow(item.GetId(), item.GetLatest(), name)
		}
		return t.Render()
	}
	return tr("No platforms matching your search.")
}

// indexesNeedUpdating returns whether one or more index files need updating.
// A duration string must be provided to calculate the time threshold
// used to update the indexes, if the duration is not valid a default
// of 24 hours is used.
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
func indexesNeedUpdating(duration string) bool {
	indexpath := configuration.DataDir(configuration.Settings)

	now := time.Now()
	modTimeThreshold, err := time.ParseDuration(duration)
	if err != nil {
		feedback.Error(tr("Invalid timeout: %s", err))
		os.Exit(errorcodes.ErrBadArgument)
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

		// should handle:
		// - package_index.json
		// - package_index.json.sig
		// - package_index.json.gz
		// - package_index.tar.bz2
		indexFileName := path.Base(URL.Path)
		indexFileName = strings.TrimSuffix(indexFileName, ".tar.bz2")
		indexFileName = strings.TrimSuffix(indexFileName, ".gz")
		indexFileName = strings.TrimSuffix(indexFileName, ".sig")
		indexFileName = strings.TrimSuffix(indexFileName, ".json")
		// and obtain package_index.json as result
		coreIndexPath := indexpath.Join(indexFileName + ".json")
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
