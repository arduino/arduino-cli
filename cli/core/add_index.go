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

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands/core"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/spf13/cobra"
)

// NewCommand creates a new `add-index` command
func initAddIndexCommand() *cobra.Command {
	addIndexCommand := &cobra.Command{
		Use:   "add-index URL ...",
		Short: "Adds URLs to user's configuration and updates indexes.",
		Long:  "Adds one or more URLs to current user's additional_urls configurations and updates indexes.",
		Example: "  " + os.Args[0] + " core add-index https://downloads.arduino.cc/packages/package_index.json\n\n" +
			"  " + os.Args[0] + " core add-index https://downloads.arduino.cc/packages/package_index.json https://dl.espressif.com/dl/package_esp32_index.json\n",
		Args: cobra.MinimumNArgs(1),
		Run:  runAddIndexCommand,
	}
	return addIndexCommand
}

func runAddIndexCommand(cmd *cobra.Command, args []string) {
	inst, err := instance.CreateInstance()
	if err != nil {
		feedback.Errorf("Error adding index urls: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	resp, err := core.AddIndex(context.Background(), &rpc.AddIndexReq{
		Instance: inst,
		IndexUrl: args,
	}, output.NewDownloadProgressBarCB())
	if err != nil {
		feedback.Errorf("Error: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	for _, message := range resp.Messages {
		feedback.Print(message)
	}
}
