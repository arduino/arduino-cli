/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package login

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/bcmi-labs/arduino-cli/auth"
	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bgentry/go-netrc/netrc"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

// Init prepares the command.
func Init(rootCommand *cobra.Command) {
	rootCommand.AddCommand(command)
}

var command = &cobra.Command{
	Use:   "login [username] [password]",
	Short: "Creates default credentials for an Arduino Create Session.",
	Long:  "Creates default credentials for an Arduino Create Session.",
	Example: "" +
		"arduino login                          # Asks for all credentials.\n" +
		"arduino login myUser MySecretPassword  # Provide all credentials.\n" +
		"arduino login myUser                   # Asks for just the password instead of having it in clear.",
	Args: cobra.RangeArgs(0, 2),
	Run:  run,
}

func run(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino login`")

	userEmpty, passwordEmpty := true, true
	if len(args) > 0 {
		userEmpty = false
		if len(args) == 2 {
			passwordEmpty = false
		}
	}
	isTextMode := formatter.IsCurrentFormat("text")
	if !isTextMode && (userEmpty || passwordEmpty) {
		formatter.PrintErrorMessage("User and password must be specified outside of text format.")
		return
	}

	var user, password string
	logrus.Info("Using/Asking credentials")
	if userEmpty {
		fmt.Print("Username: ")
		fmt.Scanln(&user)
	} else {
		user = args[0]
	}
	// Username is always lowercase.
	user = strings.ToLower(user)

	if passwordEmpty {
		fmt.Print("Password: ")
		pass, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			formatter.PrintError(err, "Cannot read password, login aborted.")
			return
		}
		password = string(pass)
		fmt.Println()
	} else {
		password = args[1]
	}

	logrus.Info("Getting ~/.netrc file")

	// Save into netrc.
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

	logrus.Info("Trying to login")

	authConf := auth.New()

	token, err := authConf.Token(user, password)
	if err != nil {
		formatter.PrintError(err, "Cannot login.")
		os.Exit(commands.ErrNetwork)
	}

	netRC.RemoveMachine("arduino.cc")
	netRC.NewMachine("arduino.cc", user, token.Access, token.Refresh)
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

	formatter.PrintResult("" +
		"Successfully logged into the system.\n" +
		"The session will continue to be refreshed with every call of the CLI and will expire if not used.")
	logrus.Info("Done")
}
