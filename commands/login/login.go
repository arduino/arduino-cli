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

package login

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/arduino/arduino-cli/auth"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/bgentry/go-netrc/netrc"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

// InitCommand prepares the command.
func InitCommand() *cobra.Command {
	loginCommand := &cobra.Command{
		Use:   "login [username] [password]",
		Short: "Creates default credentials for an Arduino Create Session.",
		Long:  "Creates default credentials for an Arduino Create Session.",
		Example: "" +
			"  " + commands.AppName + " login                          # Asks for all credentials.\n" +
			"  " + commands.AppName + " login myUser MySecretPassword  # Provide all credentials.\n" +
			"  " + commands.AppName + " login myUser                   # Asks for just the password instead of having it in clear.",
		Args: cobra.RangeArgs(0, 2),
		Run:  run,
	}
	return loginCommand
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
		pass, err := terminal.ReadPassword(syscall.Stdin)
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
