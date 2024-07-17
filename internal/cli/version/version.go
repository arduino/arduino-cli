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

package version

import (
	"context"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/updater"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewCommand created a new `version` command
func NewCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	versionCommand := &cobra.Command{
		Use:     "version",
		Short:   i18n.Tr("Shows version number of Arduino CLI."),
		Long:    i18n.Tr("Shows the version number of Arduino CLI which is installed on your system."),
		Example: "  " + os.Args[0] + " version",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runVersionCommand(cmd.Context(), srv)
		},
	}
	return versionCommand
}

func runVersionCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer) {
	logrus.Info("Executing `arduino-cli version`")

	info := version.VersionInfo
	if strings.Contains(info.VersionString, "1.0.0-snapshot") || strings.Contains(info.VersionString, "nightly") {
		// We're using a development version, no need to check if there's a
		// new release available
		feedback.PrintResult(info)
		return
	}

	latestVersion := ""
	res, err := srv.CheckForArduinoCLIUpdates(ctx, &rpc.CheckForArduinoCLIUpdatesRequest{})
	if err != nil {
		feedback.Warning("Failed to check for updates: " + err.Error())
	} else {
		latestVersion = res.GetNewestVersion()
	}

	if feedback.GetFormat() != feedback.Text {
		info.LatestVersion = latestVersion
	}

	feedback.PrintResult(info)

	if feedback.GetFormat() == feedback.Text && latestVersion != "" {
		updater.NotifyNewVersionIsAvailable(latestVersion)
	}
}
