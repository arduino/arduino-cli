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

package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/internal/cli"
	"github.com/arduino/arduino-cli/internal/cli/config"
	"github.com/arduino/arduino-cli/internal/cli/configuration"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

func main() {
	// Disable logging until it is setup in the arduino-cli pre-run
	logrus.SetOutput(io.Discard)

	// Create a new ArduinoCoreServer
	srv := commands.NewArduinoCoreServer()

	// Search for the configuration file in the command line arguments and in the environment
	configFile := configuration.FindConfigFileInArgsFallbackOnEnv(os.Args)
	ctx := config.SetConfigFile(context.Background(), configFile)

	// Read the settings from the configuration file
	openReq := &rpc.ConfigurationOpenRequest{SettingsFormat: "yaml"}
	if configData, err := paths.New(configFile).ReadFile(); err == nil {
		openReq.EncodedSettings = string(configData)
	} else if !os.IsNotExist(err) {
		feedback.FatalError(fmt.Errorf("couldn't read configuration file: %w", err), feedback.ErrGeneric)
	}
	if resp, err := srv.ConfigurationOpen(ctx, openReq); err != nil {
		feedback.FatalError(fmt.Errorf("couldn't load configuration: %w", err), feedback.ErrGeneric)
	} else if warnings := resp.GetWarnings(); len(warnings) > 0 {
		for _, warning := range warnings {
			feedback.Warning(warning)
		}
	}

	// Get the current settings from the server
	resp, err := srv.ConfigurationGet(ctx, &rpc.ConfigurationGetRequest{})
	if err != nil {
		feedback.FatalError(err, feedback.ErrGeneric)
	}
	config := resp.GetConfiguration()

	// Setup i18n
	i18n.Init(config.GetLocale())

	// Setup command line parser with the server and settings
	arduinoCmd := cli.NewCommand(srv)

	// Execute the command line
	if err := arduinoCmd.ExecuteContext(ctx); err != nil {
		feedback.FatalError(err, feedback.ErrGeneric)
	}
}
