/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package config

import (
	"net/url"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// FIXME: The way the Config objects is marshalled into JSON shouldn't be here,
// this is a temporary fix for the command `arduino-cli config dump --format json`
type jsonConfig struct {
	ProxyType           string                   `json:"proxy_type"`
	ProxyManualConfig   *jsonProxyConfig         `json:"manual_configs,omitempty"`
	SketchbookPath      string                   `json:"sketchbook_path,omitempty"`
	ArduinoDataDir      string                   `json:"arduino_data,omitempty"`
	ArduinoDownloadsDir string                   `json:"arduino_downloads_dir,omitempty"`
	BoardsManager       *jsonBoardsManagerConfig `json:"board_manager"`
}

type jsonBoardsManagerConfig struct {
	AdditionalURLS []*url.URL `json:"additional_urls,omitempty"`
}

type jsonProxyConfig struct {
	Hostname string `json:"hostname"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"` // can be encrypted, see issue #71
}

var dumpCmd = &cobra.Command{
	Use:     "dump",
	Short:   "Prints the current configuration",
	Long:    "Prints the current configuration.",
	Example: "  " + os.Args[0] + " config dump",
	Args:    cobra.NoArgs,
	Run:     runDumpCommand,
}

// ouput from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type dumpResult struct {
	structured *jsonConfig
	plain      string
}

func (dr dumpResult) Data() interface{} {
	return dr.structured
}

func (dr dumpResult) String() string {
	return dr.plain
}

func runDumpCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino config dump`")

	data, err := globals.Config.SerializeToYAML()
	if err != nil {
		feedback.Errorf("Error creating configuration: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	c := globals.Config

	sketchbookDir := ""
	if c.SketchbookDir != nil {
		sketchbookDir = c.SketchbookDir.String()
	}

	arduinoDataDir := ""
	if c.DataDir != nil {
		arduinoDataDir = c.DataDir.String()
	}

	arduinoDownloadsDir := ""
	if c.ArduinoDownloadsDir != nil {
		arduinoDownloadsDir = c.ArduinoDownloadsDir.String()
	}

	feedback.PrintResult(&dumpResult{
		structured: &jsonConfig{
			ProxyType: c.ProxyType,
			ProxyManualConfig: &jsonProxyConfig{
				Hostname: c.ProxyHostname,
				Username: c.ProxyUsername,
				Password: c.ProxyPassword,
			},
			SketchbookPath:      sketchbookDir,
			ArduinoDataDir:      arduinoDataDir,
			ArduinoDownloadsDir: arduinoDownloadsDir,
			BoardsManager: &jsonBoardsManagerConfig{
				AdditionalURLS: c.BoardManagerAdditionalUrls,
			},
		},
		plain: string(data),
	})
}
