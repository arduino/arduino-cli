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

package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func initGetCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	getCommand := &cobra.Command{
		Use:   "get",
		Short: i18n.Tr("Gets a settings key value."),
		Long:  i18n.Tr("Gets a settings key value."),
		Example: "" +
			"  " + os.Args[0] + " config get logging\n" +
			"  " + os.Args[0] + " config get daemon.port\n" +
			"  " + os.Args[0] + " config get board_manager.additional_urls",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runGetCommand(cmd.Context(), srv, args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			ctx := cmd.Context()
			return getAllSettingsKeys(ctx, srv), cobra.ShellCompDirectiveDefault
		},
	}
	return getCommand
}

func runGetCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer, args []string) {
	logrus.Info("Executing `arduino-cli config get`")

	for _, toGet := range args {
		resp, err := srv.SettingsGetValue(ctx, &rpc.SettingsGetValueRequest{Key: toGet})
		if err != nil {
			feedback.Fatal(i18n.Tr("Cannot get the configuration key %[1]s: %[2]v", toGet, err), feedback.ErrGeneric)
		}
		var result getResult
		if err := json.Unmarshal([]byte(resp.GetEncodedValue()), &result.resp); err != nil {
			// Should never happen...
			panic(fmt.Sprintf("Cannot parse JSON for key %[1]s: %[2]v", toGet, err))
		}
		feedback.PrintResult(result)
	}
}

// output from this command may require special formatting.
// create a dedicated feedback.Result implementation to safely handle
// any changes to the configuration.Settings struct.
type getResult struct {
	resp interface{}
}

func (gr getResult) Data() interface{} {
	return gr.resp
}

func (gr getResult) String() string {
	gs, err := yaml.Marshal(gr.resp)
	if err != nil {
		// Should never happen
		panic(i18n.Tr("unable to marshal config to YAML: %v", err))
	}
	return string(gs)
}
