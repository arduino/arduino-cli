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

package commands

import (
	"github.com/arduino/arduino-cli/internal/locales/cmd/commands/catalog"
	"github.com/arduino/arduino-cli/internal/locales/cmd/commands/transifex"
	"github.com/spf13/cobra"
)

var i18nCommand = &cobra.Command{
	Use:   "i18n",
	Short: "i18n",
}

func init() {
	i18nCommand.AddCommand(catalog.Command)
	i18nCommand.AddCommand(transifex.Command)
}

// Execute executes the i18n command
func Execute() error {
	return i18nCommand.Execute()
}
