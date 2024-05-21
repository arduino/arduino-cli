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
	"os"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/feedback/result"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initUpdateIndexCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	updateIndexCommand := &cobra.Command{
		Use:     "update-index",
		Short:   i18n.Tr("Updates the libraries index."),
		Long:    i18n.Tr("Updates the libraries index to the latest version."),
		Example: "  " + os.Args[0] + " lib update-index",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runUpdateIndexCommand(cmd.Context(), srv)
		},
	}
	return updateIndexCommand
}

func runUpdateIndexCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer) {
	inst := instance.CreateAndInit(ctx, srv)

	logrus.Info("Executing `arduino-cli lib update-index`")
	resp := UpdateIndex(ctx, srv, inst)
	feedback.PrintResult(&libUpdateIndexResult{result.NewUpdateLibrariesIndexResponse_ResultResult(resp)})
}

// UpdateIndex updates the index of libraries.
func UpdateIndex(ctx context.Context, srv rpc.ArduinoCoreServiceServer, inst *rpc.Instance) *rpc.UpdateLibrariesIndexResponse_Result {
	req := &rpc.UpdateLibrariesIndexRequest{Instance: inst}
	stream, resp := commands.UpdateLibrariesIndexStreamResponseToCallbackFunction(ctx, feedback.ProgressBar())
	if err := srv.UpdateLibrariesIndex(req, stream); err != nil {
		feedback.Fatal(i18n.Tr("Error updating library index: %v", err), feedback.ErrGeneric)
	}
	return resp()
}

type libUpdateIndexResult struct {
	*result.UpdateLibrariesIndexResponse_ResultResult
}

func (l *libUpdateIndexResult) String() string {
	return ""
}

func (l *libUpdateIndexResult) Data() interface{} {
	return l
}
