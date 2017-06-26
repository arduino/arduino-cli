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

	"github.com/bcmi-labs/arduino-cli/cmd/pretty_print"
	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/bcmi-labs/arduino-cli/libraries"
	"github.com/spf13/cobra"
	"github.com/zieckey/goini"
)

const (
	// LibVersion represents the `arduino lib` package version number.
	LibVersion string = "0.0.1-pre-alpha"
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
	Use:   "install [LIBRARY_NAME(S)]",
	Short: "Installs one of more specified libraries into the system.",
	Long:  `Installs one or more specified libraries into the system.`,
	RunE:  executeInstallCommand,
}

// arduinoLibUninstallCmd represents the uninstall command
var arduinoLibUninstallCmd = &cobra.Command{
	Use:   "uninstall [LIBRARY_NAME(S)]",
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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("arduino lib V. %s\n", LibVersion)
	},
}

func init() {
	arduinoCmd.AddCommand(arduinoLibCmd)
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
		fmt.Print("Cannot find index file ... ")
		err = prettyPrints.DownloadLibFileIndex().Execute(GlobalFlags.Verbose).Error
		if err != nil {
			return nil
		}
	}

	status, err := libraries.CreateStatusContextFromIndex(index)
	if err != nil {
		status, err = prettyPrints.CorruptedLibIndexFix(index, GlobalFlags.Verbose)
		if err != nil {
			return nil
		}
	}

	libraryOK := make([]string, 0, len(args))
	libraryFails := make(map[string]string, len(args))

	for i, libraryName := range args {
		library := status.Libraries[libraryName]
		if library != nil {
			//found
			_, err = libraries.DownloadAndCache(library, i+1, len(args))
			if err != nil {
				libraryFails[libraryName] = err.Error()
			} else {
				libraryOK = append(libraryOK, libraryName)
			}
		} else {
			libraryFails[libraryName] = "This library is not in library index"
		}
	}

	if GlobalFlags.Verbose > 0 {
		prettyPrints.DownloadLib(libraryOK, libraryFails)
	} else {
		for _, library := range libraryOK {
			fmt.Printf("%-10s -Downloaded\n", library)
		}
		for library, failure := range libraryFails {
			fmt.Printf("%-10s -Error : %s\n", library, failure)
		}
	}

	return nil
}

func executeInstallCommand(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("No library specified for install command")
	}

	index, err := libraries.LoadLibrariesIndex()
	if err != nil {
		fmt.Println("Cannot find index file ...")
		err = prettyPrints.DownloadLibFileIndex().Execute(GlobalFlags.Verbose).Error
		if err != nil {
			return nil
		}
	}

	status, err := libraries.CreateStatusContextFromIndex(index)
	if err != nil {
		status, err = prettyPrints.CorruptedLibIndexFix(index, GlobalFlags.Verbose)
		if err != nil {
			return nil
		}
	}

	libraryOK := make([]string, 0, len(args))
	libraryFails := make(map[string]string, len(args))

	for i, libraryName := range args {
		library := status.Libraries[libraryName]
		if library != nil {
			//found
			err = libraries.DownloadAndInstall(library, i+1, len(args))
			if err != nil {
				libraryFails[libraryName] = err.Error()
			} else {
				libraryOK = append(libraryOK, libraryName)
			}
		} else {
			libraryFails[libraryName] = "This library is not in library index"
		}
	}

	if GlobalFlags.Verbose > 0 {
		prettyPrints.InstallLib(libraryOK, libraryFails)
	} else {
		for i, library := range libraryOK {
			fmt.Printf("%-10s - Installed (%d/%d)\n", library, i+1, len(libraryOK))
		}
		i := 1
		for library, failure := range libraryFails {
			i++
			fmt.Printf("%-10s - Error : %-10s (%d/%d)\n", library, failure, i, len(libraryFails))
		}
	}

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
		fmt.Println("Cannot open libraries folder")
		return nil
	}

	libraryFails := make([]string, 0, len(args))
	libraryOK := make([]string, 0, len(args))

	dirFiles, err := dir.Readdir(0)
	if err != nil {
		fmt.Println("Cannot read into libraries folder")
		return nil
	}

	//TODO: optimize this algorithm
	// time complexity O(libraries_to_install(from RAM) *
	//                   library_folder_number(from DISK) *
	//                   library_folder_file_number (from DISK))
	//TODO : remove only one version
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
							libraryFails = append(libraryFails, library)
						} else {
							libraryOK = append(libraryOK, library)
						}
					}
				} else {
					// I use library.properties file
					content, err := os.OpenFile(indexFile, os.O_RDONLY, 0666)
					if err != nil {
						libraryFails = append(libraryFails, library)
						continue
					}

					// create map from content
					scanner := bufio.NewScanner(content)
					for scanner.Scan() {
						lines := strings.Split(scanner.Text(), "=")
						// NOTE: asserting that if there is a library.properties, there is always the
						// name of the library.
						if lines[0] == "name" {
							if strings.Contains(lines[1], library) {
								//found
								err = libraries.Uninstall(filepath.Join(libFolder, file.Name()))
								if err != nil {
									libraryFails = append(libraryFails, library)
								} else {
									libraryOK = append(libraryOK, library)
								}
							}
							break
						}
					}

					if err := scanner.Err(); err != nil {
						libraryFails = append(libraryFails, library)
					}
				}
			}
		}
	}

	if GlobalFlags.Verbose > 0 {
		if len(libraryFails) > 0 {
			fmt.Println("The following libraries were succesfully uninstalled:")
			fmt.Println(strings.Join(libraryOK, " "))
			fmt.Println("However, the uninstall process failed on the following libraries:")
			fmt.Println(strings.Join(libraryFails, " "))
		} else {
			fmt.Println("All libraries successfully uninstalled")
		}
	} else if len(libraryFails) > 0 {
		for _, failed := range libraryFails {
			fmt.Printf("%-10s - Failed\n", failed)
		}
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
		fmt.Println("Index file is corrupted. Please replace it by updating : arduino lib list update")
		return nil
	}

	status, err := libraries.CreateStatusContextFromIndex(index)
	if err != nil {
		status, err = prettyPrints.CorruptedLibIndexFix(index, GlobalFlags.Verbose)
		if err != nil {
			return nil
		}
	}

	found := false

	//Pretty print libraries from index.
	for _, name := range status.Names() {
		if strings.Contains(strings.ToLower(name), query) {
			found = true
			if GlobalFlags.Verbose > 0 {
				lib := status.Libraries[name]
				fmt.Print(lib)
				if GlobalFlags.Verbose > 1 {
					for _, r := range lib.Releases {
						fmt.Print(r)
					}
				}
				fmt.Println()
			} else {
				fmt.Println(name)
			}
		}
	}

	if !found {
		fmt.Printf("No library found matching \"%s\" search query.\n", query)
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
		fmt.Println("Cannot get libraries folder")
		return
	}

	//prettyPrints.LibStatus(status)
	dir, err := os.Open(libHome)
	if err != nil {
		fmt.Println("Cannot open libraries folder")
		return
	}

	dirFiles, err := dir.Readdir(0)
	if err != nil {
		fmt.Println("Cannot read into libraries folder")
		return
	}

	libs := make([]string, 0, 10)

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
				libs = append(libs, fileName)
			} else {
				// I use library.properties file
				content, err := ioutil.ReadFile(indexFile)
				if err != nil {
					fileName := file.Name()
					//replacing underscore in foldernames with spaces.
					fileName = strings.Replace(fileName, "_", " ", -1)
					fileName = strings.Replace(fileName, "-", " v. ", -1)
					//I use folder name
					libs = append(libs, fileName)
					continue
				}

				ini := goini.New()
				err = ini.Parse(content, "\n", "=")
				if err != nil {
					fmt.Println(err)
				}
				Name, ok := ini.Get("name")
				if !ok {
					fileName := file.Name()
					//replacing underscore in foldernames with spaces.
					fileName = strings.Replace(fileName, "_", " ", -1)
					fileName = strings.Replace(fileName, "-", " v. ", -1)
					//I use folder name
					libs = append(libs, fileName)
					continue
				}
				Version, ok := ini.Get("version")
				if !ok {
					fileName := file.Name()
					//replacing underscore in foldernames with spaces.
					fileName = strings.Replace(fileName, "_", " ", -1)
					fileName = strings.Replace(fileName, "-", " v. ", -1)
					//I use folder name
					libs = append(libs, fileName)
					continue
				}
				libs = append(libs, fmt.Sprintf("%-10s v. %s", Name, Version))
			}
		}
	}

	if len(libs) < 1 {
		fmt.Println("No library installed")
	} else {
		//pretty prints installed libraries
		for _, item := range libs {
			fmt.Println(item)
		}
	}
}

func execUpdateListIndex(cmd *cobra.Command, args []string) {
	prettyPrints.DownloadLibFileIndex().Execute(GlobalFlags.Verbose)
}
