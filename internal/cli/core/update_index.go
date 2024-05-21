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
		Short:   i18n.Tr("Updates the index of cores."),
		Long:    i18n.Tr("Updates the index of cores to the latest version."),
		Example: "  " + os.Args[0] + " core update-index",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runUpdateIndexCommand(cmd.Context(), srv)
		},
	}
	return updateIndexCommand
}

func runUpdateIndexCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer) {
	logrus.Info("Executing `arduino-cli core update-index`")
	inst := instance.CreateAndInit(ctx, srv)
	resp := UpdateIndex(ctx, srv, inst)

	feedback.PrintResult(&updateIndexResult{result.NewUpdateIndexResponse_ResultResult(resp)})
}

// UpdateIndex updates the index of platforms.
func UpdateIndex(ctx context.Context, srv rpc.ArduinoCoreServiceServer, inst *rpc.Instance) *rpc.UpdateIndexResponse_Result {
	stream, res := commands.UpdateIndexStreamResponseToCallbackFunction(ctx, feedback.ProgressBar())
	err := srv.UpdateIndex(&rpc.UpdateIndexRequest{Instance: inst}, stream)
	if err != nil {
		feedback.FatalError(err, feedback.ErrGeneric)
	}
	return res()
}

type updateIndexResult struct {
	*result.UpdateIndexResponse_ResultResult
}

func (r *updateIndexResult) Data() interface{} {
	return r
}

func (r *updateIndexResult) String() string {
	return ""
}
