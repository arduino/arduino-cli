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

package logout

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/bgentry/go-netrc/netrc"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// InitCommand prepares the command.
func InitCommand() *cobra.Command {
	logoutCommand := &cobra.Command{
		Use:     "logout",
		Short:   "Clears credentials for the Arduino Create Session.",
		Long:    "Clears credentials for the Arduino Create Session.",
		Example: "  " + commands.AppName + " logout",
		Args:    cobra.NoArgs,
		Run:     run,
	}
	return logoutCommand
}

func run(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino logout`")

	logrus.Info("Getting ~/.netrc file")
	netRCHome, err := homedir.Dir()
	if err != nil {
		formatter.PrintError(err, "Cannot get current home directory.")
		os.Exit(commands.ErrGeneric)
	}

	netRCFile := filepath.Join(netRCHome, ".netrc")
	file, err := os.OpenFile(netRCFile, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		formatter.PrintError(err, "Cannot parse .netrc file.")
		return
	}
	defer file.Close()

	netRC, err := netrc.Parse(file)
	if err != nil {
		formatter.PrintError(err, "Cannot parse .netrc file.")
		os.Exit(commands.ErrGeneric)
	}

	netRC.RemoveMachine("arduino.cc")
	content, err := netRC.MarshalText()
	if err != nil {
		formatter.PrintError(err, "Cannot parse new .netrc file.")
		os.Exit(commands.ErrGeneric)
	}

	err = ioutil.WriteFile(netRCFile, content, 0600)
	if err != nil {
		formatter.PrintError(err, "Cannot write new .netrc file.")
		os.Exit(commands.ErrGeneric)
	}

	formatter.PrintResult("Successfully logged out.")
	logrus.Info("Done")
}
