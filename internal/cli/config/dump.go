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
	"os"

	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initDumpCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	var dumpCommand = &cobra.Command{
		Use:     "dump",
		Short:   i18n.Tr("Prints the current configuration"),
		Long:    i18n.Tr("Prints the current configuration."),
		Example: "  " + os.Args[0] + " config dump",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Info("Executing `arduino-cli config dump`")
			res := &rawResult{}
			switch feedback.GetFormat() {
			case feedback.JSON, feedback.MinifiedJSON:
				resp, err := srv.ConfigurationSave(cmd.Context(), &rpc.ConfigurationSaveRequest{SettingsFormat: "json"})
				if err != nil {
					logrus.Fatalf("Error creating configuration: %v", err)
				}
				res.rawJSON = []byte(resp.GetEncodedSettings())
			case feedback.Text:
				resp, err := srv.ConfigurationSave(cmd.Context(), &rpc.ConfigurationSaveRequest{SettingsFormat: "yaml"})
				if err != nil {
					logrus.Fatalf("Error creating configuration: %v", err)
				}
				res.rawYAML = []byte(resp.GetEncodedSettings())
			default:
				logrus.Fatalf("Unsupported format: %v", feedback.GetFormat())
			}
			feedback.PrintResult(dumpResult{Config: res})
		},
	}
	return dumpCommand
}

type rawResult struct {
	rawJSON []byte
	rawYAML []byte
}

func (r *rawResult) MarshalJSON() ([]byte, error) {
	// it is already encoded in rawJSON field
	return r.rawJSON, nil
}

type dumpResult struct {
	Config *rawResult `json:"config"`
}

func (dr dumpResult) Data() interface{} {
	return dr
}

func (dr dumpResult) String() string {
	// In case of text output do not wrap the output in outer JSON or YAML structure
	return string(dr.Config.rawYAML)
}
