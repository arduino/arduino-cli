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
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
 */

package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/bcmi-labs/arduino-cli/cmd/formatter"
	"github.com/bcmi-labs/arduino-cli/cmd/output"
	"github.com/bcmi-labs/arduino-cli/cmd/pretty_print"
	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/bcmi-labs/arduino-cli/libraries"
	"github.com/bcmi-labs/arduino-cli/task"
	"github.com/spf13/cobra"
	"github.com/zieckey/goini"
	"gopkg.in/cheggaaa/pb.v1"
)

const (
	// LibVersion represents the `arduino lib` package version number.
	LibVersion string = "0.1.0-alpha.preview"
)

// arduinoLibCmd represents the libs command.
var arduinoLibCmd = &cobra.Command{
	Use:   "lib",
	Short: "Arduino commands about libraries",
	Long:  `Arduino commands about libraries`,
	Run:   executeLibCommand,
}

// arduinoLibFlags represents `arduino lib` flags.
var arduinoLibFlags struct {
	updateIndex bool
}

// arduinoLibInstallCmd represents the lib install command.
var arduinoLibInstallCmd = &cobra.Command{
	Use:   "install LIBRARY[@VERSION_NUMBER](S)",
	Short: "Installs one of more specified libraries into the system.",
	Long:  `Installs one or more specified libraries into the system.`,
	RunE:  executeInstallCommand,
}

// arduinoLibUninstallCmd represents the uninstall command
var arduinoLibUninstallCmd = &cobra.Command{
	Use:   "uninstall LIBRARY_NAME(S)",
	Short: "Uninstalls one or more libraries",
	Long:  `Uninstalls one or more libraries`,
	RunE:  executeUninstallCommand,
}

// arduinoLibSearchCmd represents the search command
var arduinoLibSearchCmd = &cobra.Command{
	Use:   "search [LIBRARY_NAME]",
	Short: "Searchs for a library data",
	Long:  `Search for one or more libraries data.`,
	RunE:  executeSearch,
}

// arduinoLibDownloadCmd represents the download command
var arduinoLibDownloadCmd = &cobra.Command{
	Use:   "download [LIBRARY_NAME(S)]",
	Short: "Downloads one or more libraries without installing them",
	Long:  `Downloads one or more libraries without installing them`,
	RunE:  executeDownloadCommand,
}

// arduinoLibListCmd represents the list libraries command.
var arduinoLibListCmd = &cobra.Command{
	Use:   "list",
	Short: "Shows a list of all installed libraries",
	Long: `Shows a list of all installed libraries.
Can be used with -v (or --verbose) flag (up to 2 times) to have longer output.`,
	Run: executeListCommand,
}

// arduinoLibVersionCmd represents the version command.
var arduinoLibVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows version Number of arduino lib",
	Long:  `Shows version Number of arduino lib which is installed on your system.`,
	Run:   executeVersionCommand,
}

func init() {
	arduinoCmd.AddCommand(arduinoLibCmd)

	versions[arduinoLibCmd.Name()] = LibVersion

	arduinoLibCmd.AddCommand(arduinoLibInstallCmd)
	arduinoLibCmd.AddCommand(arduinoLibUninstallCmd)
	arduinoLibCmd.AddCommand(arduinoLibSearchCmd)
	arduinoLibCmd.AddCommand(arduinoLibDownloadCmd)
	arduinoLibCmd.AddCommand(arduinoLibVersionCmd)
	arduinoLibCmd.AddCommand(arduinoLibListCmd)

	arduinoLibCmd.Flags().BoolVar(&arduinoLibFlags.updateIndex, "update-index", false, "Updates the libraries index")
}

func executeLibCommand(cmd *cobra.Command, args []string) {
	if arduinoLibFlags.updateIndex {
		execUpdateListIndex(cmd, args)
	} // else return
}

func executeDownloadCommand(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("No library specified for download command")
	}

	index, err := libraries.LoadLibrariesIndex()
	if err != nil {
		formatter.Print("Cannot find index file ... ")
		err = prettyPrints.DownloadLibFileIndex().Execute(GlobalFlags.Verbose).Error
		if err != nil {
			return nil
		}
	}

	status, err := index.CreateStatusContext()
	if err != nil {
		status, err = prettyPrints.CorruptedLibIndexFix(index, GlobalFlags.Verbose)
		if err != nil {
			return nil
		}
	}

	libs, failed := purgeInvalidLibraries(parseLibArgs(args), status.(libraries.StatusContext))
	libraryResults := parallelLibDownloads(libs, true, "Downloaded")

	for libFail, err := range failed {
		libraryResults[libFail] = err
	}

	formatter.Print(output.LibResultsFromMap(libraryResults))

	return nil
}

// parseLibArgs parses a sequence of "library@version" tokens and returns a map.
//
// If version is not present it is assumed as "latest" version.
func parseLibArgs(args []string) map[string]string {
	ret := make(map[string]string, len(args))
	for _, item := range args {
		tokens := strings.SplitN(item, "@", 2)
		var version string
		if len(tokens) == 2 {
			version = tokens[1]
		} else {
			version = "latest"
		}
		ret[tokens[0]] = version
	}
	return ret
}

func purgeInvalidLibraries(libnames map[string]string, status common.StatusContext) (map[*libraries.Library]string, map[string]error) {
	items := make(map[*libraries.Library]string, len(libnames))
	statusItems := status.Items()
	fails := make(map[string]error, len(libnames))

	for libraryName, version := range libnames {
		library, valid := statusItems[libraryName].(*libraries.Library)
		if !valid {
			fails[libraryName] = errors.New("Library Not Found")
		} else {
			release := library.GetVersion(version)
			if release == nil {
				fails[libraryName] = errors.New("Version Not Found")
			} else { // replaces "latest" with latest version too
				items[library] = release.Version
			}
		}
	}

	return items, fails
}

// parallelLibDownloads executes multiple libraries downloads in parallel and fills properly results.
//
// forced is used to force download if cached.
// OkStatus is used to tell the overlying process result ("Downloaded", "Installed", etc...)
func parallelLibDownloads(items map[*libraries.Library]string, forced bool, OkStatus string) map[string]interface{} {
	itemC := len(items)
	libraryResults := make(map[string]interface{}, itemC)

	tasks := make(map[string]task.Wrapper, len(items))
	progressBars := make([]*pb.ProgressBar, 0, len(items))

	textMode := formatter.IsSupported("text")

	for library, version := range items {
		release := library.GetVersion(version)
		if forced || release != nil && !library.IsCached(version) || release.CheckLocalArchive() != nil {
			var pBar *pb.ProgressBar
			if textMode {
				pBar = pb.StartNew(release.Size).SetUnits(pb.U_BYTES).Prefix(fmt.Sprintf("%-20s", library.Name))
				progressBars = append(progressBars, pBar)
			}
			tasks[library.Name] = libraries.DownloadAndCache(library, pBar, version)
		}
	}

	if len(tasks) > 0 {
		var pool *pb.Pool
		if textMode {
			pool, _ = pb.StartPool(progressBars...)
		}

		results := task.ExecuteParallelFromMap(tasks, GlobalFlags.Verbose)

		if textMode {
			pool.Stop()
		}

		for libraryName, result := range results {
			if result.Error != nil {
				libraryResults[libraryName] = result.Error
			} else {
				libraryResults[libraryName] = OkStatus
			}
		}
	}

	return libraryResults
}

func executeInstallCommand(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("No library specified for install command")
	}

	index, err := libraries.LoadLibrariesIndex()
	if err != nil {
		formatter.Print("Cannot find index file ...")
		err = prettyPrints.DownloadLibFileIndex().Execute(GlobalFlags.Verbose).Error
		if err != nil {
			return nil
		}
	}

	status, err := index.CreateStatusContext()
	if err != nil {
		status, err = prettyPrints.CorruptedLibIndexFix(index, GlobalFlags.Verbose)
		if err != nil {
			return nil
		}
	}

	libs, failed := purgeInvalidLibraries(parseLibArgs(args), status)

	libraryResults := parallelLibDownloads(libs, false, "Installed")
	for fail, reason := range failed {
		libraryResults[fail] = reason
	}
	for library, version := range libs {
		err = libraries.InstallLib(library, version)
		if err != nil {
			libraryResults[library.Name] = err
		} else {
			libraryResults[library.Name] = "Installed"
		}
	}

	formatter.Print(output.LibResultsFromMap(libraryResults))

	return nil
}

func executeUninstallCommand(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("No library specified for uninstall command")
	}

	libFolder, err := common.GetDefaultLibFolder()
	if err != nil {
		return nil
	}

	dir, err := os.Open(libFolder)
	if err != nil {
		formatter.PrintErrorMessage("Cannot open libraries folder")
		return nil
	}

	dirFiles, err := dir.Readdir(0)
	if err != nil {
		formatter.PrintErrorMessage("Cannot read into libraries folder")
		return nil
	}

	libraryResults := make(map[string]interface{}, len(args))
	for _, arg := range args {
		libraryResults[arg] = errors.New("Not Found or Not Installed")
	}
	//TODO: optimize this algorithm
	//      time complexity O(libraries_to_install(from RAM) *
	//                        library_folder_number(from DISK) *
	//                        library_folder_file_number (from DISK)).
	//TODO : remove only one version.

	for _, file := range dirFiles {
		for _, library := range args {
			if file.IsDir() {
				indexFile := filepath.Join(libFolder, file.Name(), "library.properties")
				_, err = os.Stat(indexFile)
				if os.IsNotExist(err) {
					fileName := file.Name()
					//replacing underscore in foldernames with spaces.
					fileName = strings.Replace(fileName, "_", " ", -1)
					//I use folder name
					if strings.Contains(fileName, library) {
						//found
						err = libraries.Uninstall(filepath.Join(libFolder, fileName))
						if err != nil {
							libraryResults[library] = err
						} else {
							libraryResults[library] = "Uninstalled"
						}
					}
				} else {
					// I use library.properties file
					content, err := os.OpenFile(indexFile, os.O_RDONLY, 0666)
					if err != nil {
						libraryResults[library] = err
						continue
					}

					// create map from content
					scanner := bufio.NewScanner(content)
					for scanner.Scan() {
						lines := strings.SplitN(scanner.Text(), "=", 2)
						// NOTE: asserting that if there is a library.properties, there is always the
						// name of the library.
						if lines[0] == "name" {
							if strings.Contains(lines[1], library) {
								//found
								err = libraries.Uninstall(filepath.Join(libFolder, file.Name()))
								if err != nil {
									libraryResults[library] = err
								} else {
									libraryResults[library] = "Uninstalled"
								}
							}
							break
						}
					}

					if err := scanner.Err(); err != nil {
						libraryResults[library] = err
					} else if _, ok := libraryResults[library].(error); ok {
						libraryResults[library] = errors.New("name not found in library.properties")
					}
				}
			}
		}
	}
	if len(libraryResults) > 0 {
		formatter.Print(output.LibResultsFromMap(libraryResults))
	}

	return nil
}

func executeSearch(cmd *cobra.Command, args []string) error {
	query := ""
	if len(args) > 1 {
		return errors.New("Wrong Number of Arguments")
	}
	if len(args) == 1 {
		query = strings.ToLower(strings.Join(args, " "))
	}

	index, err := libraries.LoadLibrariesIndex()
	if err != nil {
		formatter.PrintErrorMessage("Index file is corrupted. Please replace it by updating : arduino lib list update")
		return nil
	}

	status, err := index.CreateStatusContext()
	if err != nil {
		status, err = prettyPrints.CorruptedLibIndexFix(index, GlobalFlags.Verbose)
		if err != nil {
			return nil
		}
	}

	found := false

	message := output.LibSearchResults{}
	names := status.Names()
	items := status.Items()
	libs := make([]interface{}, 0, len(names))
	//Pretty print libraries from index.
	for _, name := range names {
		if strings.Contains(strings.ToLower(name), query) {
			found = true
			item := items[name].(*libraries.Library)
			if GlobalFlags.Verbose > 0 {
				libs = append(libs, item)
				if GlobalFlags.Verbose < 2 {
					item.Releases = nil
				}
			} else {
				libs = append(libs, name)
			}
		}
	}

	if !found {
		formatter.PrintErrorMessage(fmt.Sprintf("No library found matching \"%s\" search query", query))
	} else {
		message.Libraries = libs
		formatter.Print(message)
	}

	return nil
}

func executeListCommand(command *cobra.Command, args []string) {
	if arduinoLibFlags.updateIndex {
		execUpdateListIndex(command, args)
		return
	}

	libHome, err := common.GetDefaultLibFolder()
	if err != nil {
		formatter.PrintErrorMessage("Cannot get libraries folder")
		return
	}

	//prettyPrints.LibStatus(status)
	dir, err := os.Open(libHome)
	if err != nil {
		formatter.PrintErrorMessage("Cannot open libraries folder")
		return
	}

	dirFiles, err := dir.Readdir(0)
	if err != nil {
		formatter.PrintErrorMessage("Cannot read into libraries folder")
		return
	}

	libs := make(map[string]interface{}, 10)

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
				fileName := file.Name()
				//replacing underscore in foldernames with spaces.
				fileName = strings.Replace(fileName, "_", " ", -1)
				fileName = strings.Replace(fileName, "-", " v. ", -1)
				//I use folder name
				libs[fileName] = "Unknown Version"
			} else {
				// I use library.properties file
				content, err := ioutil.ReadFile(indexFile)
				if err != nil {
					fileName := file.Name()
					//replacing underscore in foldernames with spaces.
					fileName = strings.Replace(fileName, "_", " ", -1)
					fileName = strings.Replace(fileName, "-", " v. ", -1)
					//I use folder name
					libs[fileName] = "Unknown Version"
					continue
				}

				ini := goini.New()
				err = ini.Parse(content, "\n", "=")
				if err != nil {
					formatter.Print(err)
				}
				Name, ok := ini.Get("name")
				if !ok {
					fileName := file.Name()
					//replacing underscore in foldernames with spaces.
					fileName = strings.Replace(fileName, "_", " ", -1)
					fileName = strings.Replace(fileName, "-", " v. ", -1)
					//I use folder name
					libs[fileName] = "Unknown Version"
					continue
				}
				Version, ok := ini.Get("version")
				if !ok {
					fileName := file.Name()
					//replacing underscore in foldernames with spaces.
					fileName = strings.Replace(fileName, "_", " ", -1)
					fileName = strings.Replace(fileName, "-", " v. ", -1)
					//I use folder name
					libs[fileName] = "Unknown Version"
					continue
				}
				libs[Name] = fmt.Sprintf("v.%s", Version)
			}
		}
	}

	if len(libs) < 1 {
		formatter.PrintErrorMessage("No library installed")
	} else {
		formatter.Print(output.LibResultsFromMap(libs))
	}
}

func execUpdateListIndex(cmd *cobra.Command, args []string) {
	prettyPrints.DownloadLibFileIndex().Execute(GlobalFlags.Verbose)
}
