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

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/bcmi-labs/arduino-cli/cmd/formatter"
	"github.com/bcmi-labs/arduino-cli/cmd/output"
	"github.com/bcmi-labs/arduino-cli/cmd/pretty_print"
	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/bcmi-labs/arduino-cli/common/releases"
	"github.com/bcmi-labs/arduino-cli/libraries"
	"github.com/spf13/cobra"
	"github.com/zieckey/goini"
)

const (
	// LibVersion represents the `arduino lib` package version number.
	LibVersion string = "0.1.0-alpha.preview"
)

// arduinoLibCmd represents the libs command.
var arduinoLibCmd = &cobra.Command{
	Use:   "lib",
	Short: "Arduino commands about libraries",
	Long: `Arduino commands about libraries
Can be used with --update-index flag to update the libraries index too.`,
	Run: executeLibCommand,
	Example: `arduino lib install [LIBRARIES] # where 
arduino lib --update-index`,
}

// arduinoLibInstallCmd represents the lib install command.
var arduinoLibInstallCmd = &cobra.Command{
	Use:   "install LIBRARY[@VERSION_NUMBER](S)",
	Short: "Installs one of more specified libraries into the system.",
	Long:  `Installs one or more specified libraries into the system.`,
	Run:   executeInstallCommand,
	Example: `arduino lib install YoutubeApi # for the latest version
arduino lib install YoutubeApi@1.0.0     # for the specific version (in this case 1.0.0)`,
}

// arduinoLibUninstallCmd represents the uninstall command
var arduinoLibUninstallCmd = &cobra.Command{
	Use:     "uninstall LIBRARY_NAME(S)",
	Short:   "Uninstalls one or more libraries",
	Long:    `Uninstalls one or more libraries`,
	Run:     executeUninstallCommand,
	Example: ` arduino uninstall YoutubeApi`,
}

// arduinoLibSearchCmd represents the search command
var arduinoLibSearchCmd = &cobra.Command{
	Use:   "search [LIBRARY_NAME]",
	Short: "Searchs for one or more libraries data.",
	Long:  `Search for one or more libraries data (case insensitive search).`,
	Run:   executeSearchCommand,
	Example: `arduino lib search You # to show all libraries containing "You" in their name (case insensitive).
YoumadeIt
YoutubeApi`,
}

// arduinoLibDownloadCmd represents the download command
var arduinoLibDownloadCmd = &cobra.Command{
	Use:   "download [LIBRARY_NAME(S)]",
	Short: "Downloads one or more libraries without installing them",
	Long:  `Downloads one or more libraries without installing them`,
	Run:   executeDownloadCommand,
	Example: `arduino lib download YoutubeApi       # for the latest version.
arduino lib download YoutubeApi@1.0.0 # for a specific version (in this case 1.0.0)`,
}

// arduinoLibListCmd represents the list libraries command.
var arduinoLibListCmd = &cobra.Command{
	Use:   "list",
	Short: "Shows a list of all installed libraries",
	Long: `Shows a list of all installed libraries.
Can be used with -v (or --verbose) flag (up to 2 times) to have longer output.`,
	Run: executeListCommand,
	Example: `arduino lib list    # to show all installed library names
arduino lib list -v # to show more details`,
}

// arduinoLibVersionCmd represents the version command.
var arduinoLibVersionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Shows version Number of arduino lib",
	Long:    `Shows version Number of arduino lib which is installed on your system.`,
	Run:     executeVersionCommand,
	Example: arduinoVersionCmd.Example,
}

func init() {
	versions[arduinoLibCmd.Name()] = LibVersion
}

func executeLibCommand(cmd *cobra.Command, args []string) {
	if arduinoLibFlags.updateIndex {
		logrus.Info("Updating index")
		common.ExecUpdateIndex(prettyPrints.DownloadLibFileIndex())
	} else {
		logrus.Warn("Bad call, executing help command")
		cmd.Help()
		os.Exit(errBadCall)
	}
	logrus.Info("Done")
}

func executeDownloadCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino lib download`")
	if len(args) < 1 {
		logrus.Warn("No library specified for download command")
		formatter.PrintErrorMessage("No library specified for download command")
		os.Exit(errBadCall)
	}

	logrus.Info("Getting Libraries status context")
	status, err := getLibStatusContext()
	if err != nil {
		logrus.WithError(err).Warn("Cannot get status context")
		os.Exit(errGeneric)
	}

	logrus.Info("Preparing download")
	pairs := libraries.ParseArgs(args)
	libsToDownload, failOutputs := status.Process(pairs)
	outputResults := output.LibProcessResults{
		Libraries: failOutputs,
	}

	libs := make([]releases.DownloadItem, len(libsToDownload))
	for i := range libs {
		libs[i] = releases.DownloadItem(libsToDownload[i])
	}
	logrus.Info("Downloading")
	releases.ParallelDownload(libs, false, "Downloaded", &outputResults.Libraries, "library")
	logrus.Info("Download finished")
	formatter.Print(outputResults)
	logrus.Info("Done")
}

func executeInstallCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino lib install`")
	if len(args) < 1 {
		logrus.Warn("No library specified for install command")
		formatter.PrintErrorMessage("No library specified for install command")
		os.Exit(errBadCall)
	}

	logrus.Info("Getting Libraries status context")
	status, err := getLibStatusContext()
	if err != nil {
		logrus.WithError(err).Warn("Cannot get status context")
		os.Exit(errGeneric)
	}

	pairs := libraries.ParseArgs(args)
	libsToDownload, failOutputs := status.Process(pairs)
	outputResults := output.LibProcessResults{
		Libraries: failOutputs,
	}

	libs := make([]releases.DownloadItem, len(libsToDownload))
	for i := range libs {
		libs[i] = releases.DownloadItem(libsToDownload[i])
	}

	logrus.Info("Downloading")
	releases.ParallelDownload(libs, false, "Installed", &outputResults.Libraries, "library")
	logrus.Info("Download finished")

	logrus.Info("Installing")
	folder, err := common.GetDefaultLibFolder()
	if err != nil {
		logrus.WithError(err).Error("Cannot get libraries install path")
		formatter.PrintErrorMessage("Cannot get default lib install path.")
		os.Exit(errCoreConfig)
	}

	for i, item := range libsToDownload {
		err = libraries.Install(item.Name, item.Release)
		if err != nil {
			logrus.WithError(err).Warn("Library", item.Name, "errored")
			outputResults.Libraries[i] = output.ProcessResult{
				ItemName: item.Name,
				Error:    err.Error(),
			}
		} else {
			outputResults.Libraries[i].Path = filepath.Join(folder, fmt.Sprintf("%s-%s", item.Name, item.Release.VersionName()))
		}
	}

	formatter.Print(outputResults)
	logrus.Info("Done")
}

func executeUninstallCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino lib uninstall`")
	if len(args) < 1 {
		logrus.Warn("No library specified to uninstall")
		formatter.PrintErrorMessage("No library specified for uninstall command")
		os.Exit(errBadCall)
	}

	logrus.Info("Preparing")
	libs := libraries.ParseArgs(args)

	libFolder, err := common.GetDefaultLibFolder()
	if err != nil {
		logrus.WithError(err).Error("Cannot get default libraries folder")
		os.Exit(errCoreConfig)
	}

	dir, err := os.Open(libFolder)
	if err != nil {
		logrus.WithError(err).Error("Cannot open libraries folder")
		formatter.PrintErrorMessage("Cannot open libraries folder")
		os.Exit(errCoreConfig)
	}

	dirFiles, err := dir.Readdir(0)
	if err != nil {
		logrus.WithError(err).Error("Cannot read into libraries folder")
		formatter.PrintErrorMessage("Cannot read into libraries folder")
		os.Exit(errCoreConfig)
	}

	outputResults := output.LibProcessResults{
		Libraries: make([]output.ProcessResult, 0, 10),
	}

	useFileName := func(file os.FileInfo, library libraries.NameVersionPair, outputResults *output.LibProcessResults) bool {
		logrus.Info("Using file name to uninstall")
		fileName := file.Name()
		//replacing underscore in foldernames with spaces.
		fileName = strings.Replace(fileName, "_", " ", -1)
		//I use folder name
		if strings.Contains(fileName, library.Name) &&
			(library.Version == "all" || strings.Contains(fileName, library.Version)) {
			result := output.ProcessResult{
				ItemName: fmt.Sprint(library.Name, "@", library.Version),
			}
			//found
			err = libraries.Uninstall(filepath.Join(libFolder, fileName))
			if err != nil {
				logrus.WithError(err).Warn("Cannot uninstall", fileName)
				result.Error = err.Error()
				(*outputResults).Libraries = append((*outputResults).Libraries, result)
			} else {
				logrus.Info(fileName, "Uninstalled")
				result.Error = "Uninstalled"
				(*outputResults).Libraries = append((*outputResults).Libraries, result)
			}
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
		//readapting "latest" to "any" to avoid to use two struct with a minor change.
		if library.Version == "latest" {
			library.Version = "all"
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
					// I use library.properties file
					content, err := ioutil.ReadFile(indexFile)
					if err != nil {
						logrus.WithError(err).Warn("Cannot read library.properties")
						outputResults.Libraries = append(outputResults.Libraries, output.ProcessResult{
							ItemName: fmt.Sprint(library.Name, "@", library.Version),
							Error:    err.Error(),
						})
						break
					}

					logrus.Info("Parsing library.properties")
					ini := goini.New()
					err = ini.Parse(content, "\n", "=")
					if err != nil {
						logrus.WithError(err).Warn("Cannot parse library.properties")
						formatter.Print(err)
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
						(library.Version == "all" || library.Version == version) {
						logrus.Info("Uninstalling", file.Name())
						err := libraries.Uninstall(filepath.Join(libFolder, file.Name()))
						if err != nil {
							logrus.WithError(err).Warn("Cannot uninstall", file.Name())
						}
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

func executeSearchCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino lib search`")
	query := strings.ToLower(strings.Join(args, " "))

	logrus.Info("Getting libraries status context")
	status, err := getLibStatusContext()
	if err != nil {
		logrus.WithError(err).Error("Cannot get libraries status context")
		os.Exit(errCoreConfig)
	}

	logrus.Info("Preparing")

	found := false
	names := status.Names()
	message := output.LibSearchResults{
		Libraries: make([]interface{}, 0, len(names)),
	}

	logrus.Info("Searching")
	items := status.Libraries
	//Pretty print libraries from index.
	for _, name := range names {
		if strings.Contains(strings.ToLower(name), query) {
			found = true
			if arduinoLibSearchFlags.Names {
				if formatter.IsCurrentFormat("text") {
					name = fmt.Sprintf("\"%s\"", name)
				}
				message.Libraries = append(message.Libraries, name)
			} else {
				message.Libraries = append(message.Libraries, items[name])
			}
		}
	}

	if !found {
		logrus.Warnf("No library found matching \"%s\" search query", query)
		formatter.PrintErrorMessage(fmt.Sprintf("No library found matching \"%s\" search query", query))
	} else {
		formatter.Print(message)
	}
	logrus.Info("Done")
}

func executeListCommand(command *cobra.Command, args []string) {
	logrus.Info("Executing `arduino lib list`")

	libHome, err := common.GetDefaultLibFolder()
	if err != nil {
		logrus.WithError(err).Error("Cannot get libraries folder")
		formatter.PrintErrorMessage("Cannot get libraries folder")
		os.Exit(errCoreConfig)
	}

	dir, err := os.Open(libHome)
	if err != nil {
		logrus.WithError(err).Error("Cannot open libraries folder")
		formatter.PrintErrorMessage("Cannot open libraries folder")
		os.Exit(errCoreConfig)
	}
	defer dir.Close()

	dirFiles, err := dir.Readdir(0)
	if err != nil {
		logrus.WithError(err).Error("Cannot read into libraries folder")
		formatter.PrintErrorMessage("Cannot read into libraries folder")
		os.Exit(errCoreConfig)
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
					formatter.Print(err)
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
		formatter.PrintErrorMessage("No library installed")
	} else {
		formatter.Print(libs)
	}
	logrus.Info("Done")
}

func resultFromFileName(file os.FileInfo, libs *output.LibProcessResults) {
	fileName := file.Name()
	//replacing underscore in foldernames with spaces.
	fileName = strings.Replace(fileName, "_", " ", -1)
	fileName = strings.Replace(fileName, "-", " v. ", -1)
	//I use folder name
	logrus.WithField("Name", fileName).Warn("Using filename to get result")
	libs.Libraries = append(libs.Libraries, output.ProcessResult{
		ItemName: fileName,
		Status:   "",
		Error:    "Unknown Version",
	})
}

func getLibStatusContext() (*libraries.StatusContext, error) {
	var index libraries.Index
	err := libraries.LoadIndex(&index)
	if err != nil {
		logrus.WithError(err).Warn("Error during index load, pretty printing error message and trying to recover")
		status, err := prettyPrints.CorruptedLibIndexFix(index)
		if err != nil {
			logrus.WithError(err).Error("Did not recover, returning error")
			return nil, err
		}
		logrus.Warn("Recovered and status context created")
		return &status, nil
	}

	logrus.Info("Creating status context")
	status := index.CreateStatusContext()
	return &status, nil
}
