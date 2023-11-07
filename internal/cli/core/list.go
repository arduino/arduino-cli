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
	"os"

	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/feedback/result"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/table"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initListCommand() *cobra.Command {
	var updatableOnly bool
	var all bool
	listCommand := &cobra.Command{
		Use:     "list",
		Short:   tr("Shows the list of installed platforms."),
		Long:    tr("Shows the list of installed platforms."),
		Example: "  " + os.Args[0] + " core list",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runListCommand(args, all, updatableOnly)
		},
	}
	listCommand.Flags().BoolVar(&updatableOnly, "updatable", false, tr("List updatable platforms."))
	listCommand.Flags().BoolVar(&all, "all", false, tr("If set return all installable and installed cores, including manually installed."))
	return listCommand
}

func runListCommand(args []string, all bool, updatableOnly bool) {
	inst := instance.CreateAndInit()
	logrus.Info("Executing `arduino-cli core list`")
	List(inst, all, updatableOnly)
}

// List gets and prints a list of installed platforms.
func List(inst *rpc.Instance, all bool, updatableOnly bool) {
	platforms := GetList(inst, all, updatableOnly)
	feedback.PrintResult(newCoreListResult(platforms, updatableOnly))
}

// GetList returns a list of installed platforms.
func GetList(inst *rpc.Instance, all bool, updatableOnly bool) []*rpc.PlatformSummary {
	platforms, err := core.PlatformSearch(&rpc.PlatformSearchRequest{
		Instance:          inst,
		ManuallyInstalled: true,
	})
	if err != nil {
		feedback.Fatal(tr("Error listing platforms: %v", err), feedback.ErrGeneric)
	}

	// If both `all` and `updatableOnly` are set, `all` takes precedence.
	if all {
		return platforms.GetSearchOutput()
	}

	result := []*rpc.PlatformSummary{}
	for _, platform := range platforms.GetSearchOutput() {
		if platform.InstalledVersion == "" && !platform.GetMetadata().ManuallyInstalled {
			continue
		}
		if updatableOnly && platform.InstalledVersion == platform.LatestVersion {
			continue
		}
		result = append(result, platform)
	}
	return result
}

func newCoreListResult(in []*rpc.PlatformSummary, updatableOnly bool) *coreListResult {
	res := &coreListResult{updatableOnly: updatableOnly}
	for _, platformSummary := range in {
		res.Platforms = append(res.Platforms, result.NewPlatformSummary(platformSummary))
	}
	return res
}

type coreListResult struct {
	Platforms     []*result.PlatformSummary `json:"platforms"`
	updatableOnly bool
}

// Data implements Result interface
func (ir coreListResult) Data() interface{} {
	return ir
}

// String implements Result interface
func (ir coreListResult) String() string {
	if len(ir.Platforms) == 0 {
		if ir.updatableOnly {
			return tr("All platforms are up to date.")
		}
		return tr("No platforms installed.")
	}
	t := table.New()
	t.SetHeader(tr("ID"), tr("Installed"), tr("Latest"), tr("Name"))
	for _, platform := range ir.Platforms {
		latestVersion := platform.LatestVersion.String()
		if latestVersion == "" {
			latestVersion = "n/a"
		}
		t.AddRow(platform.Id, platform.InstalledVersion, latestVersion, platform.GetPlatformName())
	}

	return t.Render()
}
