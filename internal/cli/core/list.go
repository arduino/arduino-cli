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
	"os"

	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/feedback/result"
	"github.com/arduino/arduino-cli/internal/cli/feedback/table"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initListCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	var updatableOnly bool
	var all bool
	listCommand := &cobra.Command{
		Use:     "list",
		Short:   i18n.Tr("Shows the list of installed platforms."),
		Long:    i18n.Tr("Shows the list of installed platforms."),
		Example: "  " + os.Args[0] + " core list",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runListCommand(cmd.Context(), srv, all, updatableOnly)
		},
	}
	listCommand.Flags().BoolVar(&updatableOnly, "updatable", false, i18n.Tr("List updatable platforms."))
	listCommand.Flags().BoolVar(&all, "all", false, i18n.Tr("If set return all installable and installed cores, including manually installed."))
	return listCommand
}

func runListCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer, all bool, updatableOnly bool) {
	inst := instance.CreateAndInit(ctx, srv)
	logrus.Info("Executing `arduino-cli core list`")
	List(ctx, srv, inst, all, updatableOnly)
}

// List gets and prints a list of installed platforms.
func List(ctx context.Context, srv rpc.ArduinoCoreServiceServer, inst *rpc.Instance, all bool, updatableOnly bool) {
	platforms := GetList(ctx, srv, inst, all, updatableOnly)
	feedback.PrintResult(newCoreListResult(platforms, updatableOnly))
}

// GetList returns a list of installed platforms.
func GetList(ctx context.Context, srv rpc.ArduinoCoreServiceServer, inst *rpc.Instance, all bool, updatableOnly bool) []*rpc.PlatformSummary {
	platforms, err := srv.PlatformSearch(ctx, &rpc.PlatformSearchRequest{
		Instance:          inst,
		ManuallyInstalled: true,
	})
	if err != nil {
		feedback.Fatal(i18n.Tr("Error listing platforms: %v", err), feedback.ErrGeneric)
	}

	// If both `all` and `updatableOnly` are set, `all` takes precedence.
	if all {
		return platforms.GetSearchOutput()
	}

	result := []*rpc.PlatformSummary{}
	for _, platform := range platforms.GetSearchOutput() {
		if platform.GetInstalledVersion() == "" && !platform.GetMetadata().GetManuallyInstalled() {
			continue
		}
		if updatableOnly && platform.GetInstalledVersion() == platform.GetLatestVersion() {
			continue
		}
		result = append(result, platform)
	}
	return result
}

func newCoreListResult(in []*rpc.PlatformSummary, updatableOnly bool) *coreListResult {
	res := &coreListResult{updatableOnly: updatableOnly, Platforms: make([]*result.PlatformSummary, len(in))}
	for i, platformSummary := range in {
		res.Platforms[i] = result.NewPlatformSummary(platformSummary)
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
			return i18n.Tr("All platforms are up to date.")
		}
		return i18n.Tr("No platforms installed.")
	}
	t := table.New()
	t.SetHeader(i18n.Tr("ID"), i18n.Tr("Installed"), i18n.Tr("Latest"), i18n.Tr("Name"))
	for _, platform := range ir.Platforms {
		latestVersion := platform.LatestVersion.String()
		if latestVersion == "" {
			latestVersion = "n/a"
		}
		t.AddRow(platform.Id, platform.InstalledVersion, latestVersion, platform.GetPlatformName())
	}

	return t.Render()
}
