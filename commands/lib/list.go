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

package lib

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zieckey/goini"
)

func init() {
	command.AddCommand(listCommand)
}

var listCommand = &cobra.Command{
	Use:   "list",
	Short: "Shows a list of all installed libraries.",
	Long: "Shows a list of all installed libraries.\n" +
		"Can be used with -v (or --verbose) flag (up to 2 times) to have longer output.",
	Example: "" +
		"arduino lib list    # to show all installed library names.\n" +
		"arduino lib list -v # to show more details.",
	Args: cobra.NoArgs,
	Run:  runListCommand,
}

func runListCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino lib list`")

	libHome, err := configs.LibrariesFolder.Get()
	if err != nil {
		formatter.PrintError(err, "Cannot get libraries folder.")
		os.Exit(commands.ErrCoreConfig)
	}

	dir, err := os.Open(libHome)
	if err != nil {
		formatter.PrintError(err, "Cannot open libraries folder.")
		os.Exit(commands.ErrCoreConfig)
	}
	defer dir.Close()

	dirFiles, err := dir.Readdir(0)
	if err != nil {
		formatter.PrintError(err, "Cannot read into libraries folder.")
		os.Exit(commands.ErrCoreConfig)
	}

	libs := output.LibProcessResults{
		Libraries: make([]output.ProcessResult, 0, 10),
	}

	logrus.Info("Listing")

	//TODO: optimize this algorithm
	// time complexity O(libraries_to_install(from RAM) *
	//                   library_folder_number(from DISK) *
	//                   library_folder_file_number (from DISK))
	//TODO : remove only one version
	for _, file := range dirFiles {
		if file.IsDir() {
			indexFile := filepath.Join(libHome, file.Name(), "library.properties")
			_, err = os.Stat(indexFile)
			if os.IsNotExist(err) {
				logrus.WithError(err).Warn("No library.properties for this library")
				resultFromFileName(file, &libs)
			} else {
				logrus.Infof("Using library.properties for %s", file.Name())
				content, err := ioutil.ReadFile(indexFile)
				if err != nil {
					logrus.WithError(err).Warn("Cannot read into library.properties for this library")
					resultFromFileName(file, &libs)
					continue
				}

				logrus.Info("Parsing library.properties")

				ini := goini.New()
				err = ini.Parse(content, "\n", "=")
				if err != nil {
					logrus.WithError(err).Warn("Cannot parse library.properties")
					resultFromFileName(file, &libs)
					continue
				}
				Name, ok := ini.Get("name")
				if !ok {
					logrus.Warn("Name not found in library.properties")
					resultFromFileName(file, &libs)
					continue
				}
				Version, ok := ini.Get("version")
				if !ok {
					logrus.Warn("Version not found in library.properties")
					resultFromFileName(file, &libs)
					continue
				}
				libs.Libraries = append(libs.Libraries, output.ProcessResult{
					ItemName: Name,
					Status:   fmt.Sprint("v.", Version),
					Error:    "",
				})
			}
		}
	}

	if len(libs.Libraries) < 1 {
		formatter.PrintErrorMessage("No library installed.")
	} else {
		formatter.Print(libs)
	}
	logrus.Info("Done")
}
