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

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initDownloadCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	downloadCommand := &cobra.Command{
		Use:   fmt.Sprintf("download [%s:%s[@%s]]...", i18n.Tr("PACKAGER"), i18n.Tr("ARCH"), i18n.Tr("VERSION")),
		Short: i18n.Tr("Downloads one or more cores and corresponding tool dependencies."),
		Long:  i18n.Tr("Downloads one or more cores and corresponding tool dependencies."),
		Example: "" +
			"  " + os.Args[0] + " core download arduino:samd       # " + i18n.Tr("download the latest version of Arduino SAMD core.") + "\n" +
			"  " + os.Args[0] + " core download arduino:samd@1.6.9 # " + i18n.Tr("download a specific version (in this case 1.6.9)."),
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runDownloadCommand(cmd.Context(), srv, args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return arguments.GetInstallableCores(cmd.Context(), srv), cobra.ShellCompDirectiveDefault
		},
	}
	return downloadCommand
}

func runDownloadCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer, args []string) {
	inst := instance.CreateAndInit(ctx, srv)

	logrus.Info("Executing `arduino-cli core download`")

	platformsRefs, err := arguments.ParseReferences(ctx, srv, args)
	if err != nil {
		feedback.Fatal(i18n.Tr("Invalid argument passed: %v", err), feedback.ErrBadArgument)
	}

	for i, platformRef := range platformsRefs {
		platformDownloadreq := &rpc.PlatformDownloadRequest{
			Instance:        inst,
			PlatformPackage: platformRef.PackageName,
			Architecture:    platformRef.Architecture,
			Version:         platformRef.Version,
		}
		stream := commands.PlatformDownloadStreamResponseToCallbackFunction(ctx, feedback.ProgressBar())
		if err := srv.PlatformDownload(platformDownloadreq, stream); err != nil {
			feedback.Fatal(i18n.Tr("Error downloading %[1]s: %[2]v", args[i], err), feedback.ErrNetwork)
		}
	}
}
