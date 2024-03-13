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
	"strings"
	"time"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/feedback/result"
	"github.com/arduino/arduino-cli/internal/cli/feedback/table"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initSearchCommand() *cobra.Command {
	var allVersions bool
	searchCommand := &cobra.Command{
		Use:     fmt.Sprintf("search <%s...>", tr("keywords")),
		Short:   tr("Search for a core in Boards Manager."),
		Long:    tr("Search for a core in Boards Manager using the specified keywords."),
		Example: "  " + os.Args[0] + " core search MKRZero -a -v",
		Args:    cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runSearchCommand(cmd, args, allVersions)
		},
	}
	searchCommand.Flags().BoolVarP(&allVersions, "all", "a", false, tr("Show all available core versions."))

	return searchCommand
}

// indexUpdateInterval specifies the time threshold over which indexes are updated
const indexUpdateInterval = 24 * time.Hour

func runSearchCommand(cmd *cobra.Command, args []string, allVersions bool) {
	inst := instance.CreateAndInit()

	res, err := commands.UpdateIndex(
		context.Background(),
		&rpc.UpdateIndexRequest{Instance: inst, UpdateIfOlderThanSecs: int64(indexUpdateInterval.Seconds())},
		feedback.ProgressBar())
	if err != nil {
		feedback.FatalError(err, feedback.ErrGeneric)
	}
	for _, idxRes := range res.GetUpdatedIndexes() {
		if idxRes.GetStatus() == rpc.IndexUpdateReport_STATUS_UPDATED {
			// At least one index has been updated, reinitialize the instance
			instance.Init(inst)
			break
		}
	}

	arguments := strings.ToLower(strings.Join(args, " "))
	logrus.Infof("Executing `arduino-cli core search` with args: '%s'", arguments)

	resp, err := commands.PlatformSearch(&rpc.PlatformSearchRequest{
		Instance:   inst,
		SearchArgs: arguments,
	})
	if err != nil {
		feedback.Fatal(tr("Error searching for platforms: %v", err), feedback.ErrGeneric)
	}

	coreslist := resp.GetSearchOutput()
	feedback.PrintResult(newSearchResult(coreslist, allVersions))
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type searchResults struct {
	Platforms   []*result.PlatformSummary `json:"platforms"`
	allVersions bool
}

func newSearchResult(in []*rpc.PlatformSummary, allVersions bool) *searchResults {
	res := &searchResults{
		Platforms:   make([]*result.PlatformSummary, len(in)),
		allVersions: allVersions,
	}
	for i, platformSummary := range in {
		res.Platforms[i] = result.NewPlatformSummary(platformSummary)
	}
	return res
}

func (sr searchResults) Data() interface{} {
	return sr
}

func (sr searchResults) String() string {
	if len(sr.Platforms) == 0 {
		return tr("No platforms matching your search.")
	}

	t := table.New()
	t.SetHeader(tr("ID"), tr("Version"), tr("Name"))

	addRow := func(platform *result.PlatformSummary, release *result.PlatformRelease) {
		if release == nil {
			t.AddRow(platform.Id, "n/a", platform.GetPlatformName())
			return
		}
		t.AddRow(platform.Id, release.Version, release.FormatName())
	}

	for _, platform := range sr.Platforms {
		// When allVersions is not requested we only show the latest compatible version
		if !sr.allVersions {
			addRow(platform, platform.GetLatestRelease())
			continue
		}

		for _, release := range platform.Releases.Values() {
			addRow(platform, release)
		}
	}
	return t.Render()
}
