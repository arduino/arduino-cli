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

package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/bcmi-labs/arduino-cli/cores"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Init prepares the command.
func Init(rootCommand *cobra.Command) {
	rootCommand.AddCommand(command)
}

var command = &cobra.Command{
	Use:     "validate",
	Short:   "Validates Arduino installation.",
	Long:    "Checks installed cores and tools for corruption.",
	Example: "arduino validate",
	Args:    cobra.NoArgs,
	Run:     run,
}

func run(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino validate`")
	packagesFolder, err := configs.PackagesFolder.Get()
	if err != nil {
		formatter.PrintError(err, "Cannot get packages folder.")
		os.Exit(commands.ErrCoreConfig)
	}
	err = filepath.Walk(packagesFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		relativePath, err := filepath.Rel(packagesFolder, path)
		if err != nil {
			return err
		}
		pathParts := strings.Split(relativePath, string(filepath.Separator))
		if len(pathParts) == 4 {
			isValid, err := cores.CheckDirChecksum(path)
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
