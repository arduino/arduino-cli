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
	command.Flags().StringVarP(&flags.user, "user", "u", "", "The username used to log in.")
	command.Flags().StringVarP(&flags.password, "password", "p", "", "The password used to authenticate.")
}

var flags struct {
	user     string // The user who asks to login.
	password string // The password used to authenticate.
}

var command = &cobra.Command{
	Use:   "login [--user USER --password PASSWORD | --user USER",
	Short: "create default credentials for an Arduino Create Session.",
	Long:  "create default credentials for an Arduino Create Session.",
	Example: "" +
		"arduino login                          # Asks all credentials.\n" +
		"arduino login --user myUser --password MySecretPassword\n" +
		"arduino login --user myUser --password # Asks to write the password inside the command instead of having it in clear.",
	Run: run,
}

func run(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino login`")

	userEmpty := flags.user == ""
	passwordEmpty := flags.password == ""
	isTextMode := formatter.IsCurrentFormat("text")
	if !isTextMode && (userEmpty || passwordEmpty) {
		formatter.PrintErrorMessage("User and password must be specified outside of text format.")
		return
	}

	logrus.Info("Using/Asking credentials")
	if userEmpty {
		fmt.Print("Username: ")
		fmt.Scanln(&flags.user)
	}

	if passwordEmpty {
		fmt.Print("Password: ")
		pass, err := terminal.ReadPassword(syscall.Stdin)
		if err != nil {
			formatter.PrintError(err, "Cannot read password, login aborted.")
			return
		}
		flags.password = string(pass)
		fmt.Println()
	}

	logrus.Info("Getting ~/.netrc file")

	//save into netrc
	netRCHome, err := homedir.Dir()
	if err != nil {
		formatter.PrintError(err, "Cannot get current home directory.")
		os.Exit(commands.ErrGeneric)
	}

	netRCFile := filepath.Join(netRCHome, ".netrc")
	file, err := os.OpenFile(netRCFile, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		formatter.PrintError(err, "Cannot parse .netrc file.")
		return
	}
	defer file.Close()
	NetRC, err := netrc.Parse(file)
	if err != nil {
		formatter.PrintError(err, "Cannot parse .netrc file.")
		os.Exit(commands.ErrGeneric)
	}

	logrus.Info("Trying to login")

	usr := flags.user
	pwd := flags.password
	authConf := auth.New()

	token, err := authConf.Token(usr, pwd)
	if err != nil {
		formatter.PrintError(err, "Cannot login.")
		os.Exit(commands.ErrNetwork)
	}

	NetRC.RemoveMachine("arduino.cc")
	NetRC.NewMachine("arduino.cc", usr, token.Access, token.Refresh)
	content, err := NetRC.MarshalText()
	if err != nil {
		formatter.PrintError(err, "Cannot parse new .netrc file.")
		os.Exit(commands.ErrGeneric)
	}

	err = ioutil.WriteFile(netRCFile, content, 0666)
	if err != nil {
		formatter.PrintError(err, "Cannot write new .netrc file.")
		os.Exit(commands.ErrGeneric)
	}

	formatter.PrintResult("" +
		"Successfully logged into the system.\n" +
		"The session will continue to be refreshed with every call of the CLI and will expire if not used.")
	logrus.Info("Done")
}
