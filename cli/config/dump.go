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

	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func initDumpCommand() *cobra.Command {
	var dumpCommand = &cobra.Command{
		Use:     "dump",
		Short:   tr("Prints the current configuration"),
		Long:    tr("Prints the current configuration."),
		Example: "  " + os.Args[0] + " config dump",
		Args:    cobra.NoArgs,
		Run:     runDumpCommand,
	}
	return dumpCommand
}

func runDumpCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino-cli config dump`")
	feedback.PrintResult(dumpResult{configuration.Settings.AllSettings()})
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type dumpResult struct {
	data map[string]interface{}
}

func (dr dumpResult) Data() interface{} {
	return dr.data
}

func (dr dumpResult) String() string {
	bs, err := yaml.Marshal(dr.data)
	if err != nil {
		// Should never happen
		panic(tr("unable to marshal config to YAML: %v", err))
	}
	return string(bs)
}
