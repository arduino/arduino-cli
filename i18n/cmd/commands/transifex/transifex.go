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

package transifex

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Command is the transifex command
var Command = &cobra.Command{
	Use:              "transifex",
	Short:            "transifex",
	PersistentPreRun: preRun,
}

var project string
var resource string
var apiKey string

func init() {
	Command.AddCommand(pullTransifexCommand)
	Command.AddCommand(pushTransifexCommand)
}

func preRun(cmd *cobra.Command, args []string) {
	project = os.Getenv("TRANSIFEX_PROJECT")
	resource = os.Getenv("TRANSIFEX_RESOURCE")
	apiKey = os.Getenv("TRANSIFEX_RESOURCE")

	if project = os.Getenv("TRANSIFEX_PROJECT"); project == "" {
		fmt.Println("missing TRANSIFEX_PROJECT environment variable")
		os.Exit(1)
	}

	if resource = os.Getenv("TRANSIFEX_RESOURCE"); resource == "" {
		fmt.Println("missing TRANSIFEX_RESOURCE environment variable")
		os.Exit(1)
	}

	if apiKey = os.Getenv("TRANSIFEX_API_KEY"); apiKey == "" {
		fmt.Println("missing TRANSIFEX_API_KEY environment variable")
		os.Exit(1)
	}

	return
}
