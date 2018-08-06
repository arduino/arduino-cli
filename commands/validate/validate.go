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

package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bcmi-labs/arduino-cli/arduino/resources"
	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// InitCommand prepares the command.
func InitCommand() *cobra.Command {
	var validateCommand = &cobra.Command{
		Use:     "validate",
		Short:   "Validates Arduino installation.",
		Long:    "Checks installed cores and tools for corruption.",
		Example: "arduino validate",
		Args:    cobra.NoArgs,
		Run:     run,
	}
	return validateCommand
}

func run(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino validate`")
	packagesDir := commands.Config.PackagesDir().String()
	err := filepath.Walk(packagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		relativePath, err := filepath.Rel(packagesDir, path)
		if err != nil {
			return err
		}
		pathParts := strings.Split(relativePath, string(filepath.Separator))
		if len(pathParts) == 4 {
			isValid, err := resources.CheckDirChecksum(path)
			if err != nil {
				return err
			}
			if !isValid {
				formatter.PrintErrorMessage(fmt.Sprintf("Corrupted %s", path))
			}
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		formatter.PrintError(err, "Failed to perform validation.")
		os.Exit(commands.ErrBadCall)
	}
}
