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
	"strings"

	"github.com/bcmi-labs/arduino-cli/arduino/libraries"
	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zieckey/goini"
)

const (
	versionAll    string = "all"
	versionLatest string = "latest"
)

func initUninstallCommand() *cobra.Command {
	uninstallCommand := &cobra.Command{
		Use:     "uninstall LIBRARY_NAME(S)",
		Short:   "Uninstalls one or more libraries.",
		Long:    "Uninstalls one or more libraries.",
		Example: "arduino lib uninstall AudioZero",
		Args:    cobra.MinimumNArgs(1),
		Run:     runUninstallCommand,
	}
	return uninstallCommand
}

func runUninstallCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino lib uninstall`")

	logrus.Info("Preparing")
	libs := libraries.ParseArgs(args)

	libFolder, err := configs.LibrariesFolder.Get()
	if err != nil {
		formatter.PrintError(err, "Cannot get default libraries folder.")
		os.Exit(commands.ErrCoreConfig)
	}

	dir, err := os.Open(libFolder)
	if err != nil {
		formatter.PrintError(err, "Cannot open libraries folder.")
		os.Exit(commands.ErrCoreConfig)
	}

	dirFiles, err := dir.Readdir(0)
	if err != nil {
		formatter.PrintError(err, "Cannot read into libraries folder.")
		os.Exit(commands.ErrCoreConfig)
	}

	outputResults := output.LibProcessResults{
		Libraries: map[string]output.ProcessResult{},
	}

	useFileName := func(file os.FileInfo, library libraries.Reference, outputResults *output.LibProcessResults) bool {
		logrus.Info("Using file name to uninstall")
		fileName := file.Name()
		// Replacing underscore in foldernames with spaces.
		fileName = strings.Replace(fileName, "_", " ", -1)
		// I use folder name.
		if strings.Contains(fileName, library.Name) &&
			(library.Version == versionAll || strings.Contains(fileName, library.Version)) {
			result := output.ProcessResult{
				ItemName: fmt.Sprint(library.Name, "@", library.Version),
			}
			// Found.
			err = libraries.Uninstall(filepath.Join(libFolder, fileName))
			if err != nil {
				logrus.WithError(err).Warn("Cannot uninstall", fileName)
				result.Error = err.Error()
			} else {
				logrus.Info(fileName, "Uninstalled")
				result.Error = "Uninstalled"
			}
			// FIXME: Should use GetLibraryCode but we don't have a damn library here -.-'
			(*outputResults).Libraries[library.Name] = result
			return true
		}
		return false
	}

	logrus.Info("Removing libraries")

	//TODO: optimize this algorithm
	//      time complexity O(libraries_to_install(from RAM) *
	//                        library_folder_number(from DISK) *
	//                        library_folder_file_number (from DISK)).
	for _, library := range libs {
		// Readapting "latest" to "any" to avoid to use two struct with a minor change.
		if library.Version == versionLatest {
			library.Version = versionAll
		}
		for _, file := range dirFiles {
			if file.IsDir() {
				indexFile := filepath.Join(libFolder, file.Name(), "library.properties")
				_, err = os.Stat(indexFile)
				if os.IsNotExist(err) {
					if useFileName(file, library, &outputResults) {
						break
					}
				} else if err == nil {
					logrus.Info("using library.properties for", library.Name)
					// I use library.properties file.
					content, err := ioutil.ReadFile(indexFile)
					if err != nil {
						logrus.WithError(err).Warn("Cannot read library.properties")
						// FIXME: Should use GetLibraryCode but we don't have a damn library here -.-'
						outputResults.Libraries[library.Name] = output.ProcessResult{
							ItemName: fmt.Sprint(library.Name, "@", library.Version),
							Error:    err.Error(),
						}
						break
					}

					logrus.Info("Parsing library.properties")
					ini := goini.New()
					err = ini.Parse(content, "\n", "=")
					if err != nil {
						logrus.WithError(err).Warn("Cannot parse library.properties")
						if useFileName(file, library, &outputResults) {
							break
						}
						continue
					}
					name, ok := ini.Get("name")
					if !ok {
						logrus.Warn("Name not found in library.properties")
						if useFileName(file, library, &outputResults) {
							break
						}
						continue
					}
					version, ok := ini.Get("version")
					if !ok {
						logrus.Warn("Version not found in library.properties")
						if useFileName(file, library, &outputResults) {
							break
						}
						continue
					}
					if name == library.Name &&
						(library.Version == versionAll || library.Version == version) {
						logrus.Info("Uninstalling ", file.Name())
						err := libraries.Uninstall(filepath.Join(libFolder, file.Name()))
						result := output.ProcessResult{
							ItemName: fmt.Sprint(library.Name, "@", library.Version),
						}
						if err != nil {
							logrus.WithError(err).Warn("Cannot uninstall", file.Name())
							result.Error = err.Error()
						} else {
							logrus.Info(file.Name(), "Uninstalled")
							result.Error = "Uninstalled"
						}
						// FIXME: Should use GetLibraryCode but we don't have a damn library here -.-'
						outputResults.Libraries[library.Name] = result
					}
				}
			}
		}
	}

	if len(outputResults.Libraries) > 0 {
		formatter.Print(outputResults)
	}
	logrus.Info("Done")
}
