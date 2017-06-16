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
	"os"
	"path/filepath"
	"strings"

	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/bcmi-labs/arduino-cli/libraries"
	"github.com/spf13/cobra"
)

const (
	// LibVersion represents the `arduino lib` package version number.
	LibVersion string = "0.0.1-pre-alpha"
)

// arduinoLibCmd represents the libs command
var arduinoLibCmd = &cobra.Command{
	Use:   "lib",
	Short: "Shows all commands regarding libraries.",
	Long:  `Shows all commands regarding libraries.`,
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
	Use:   "uninstall",
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
}

func executeDownloadCommand(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("No library specified for download command")
	}

	index, err := libraries.LoadLibrariesIndex()
	if err != nil {
		fmt.Print("Cannot find index file ... ")
		err = prettyPrintDownloadFileIndex()
		if err != nil {
			return nil
		}
	}

	status, err := libraries.CreateStatusContextFromIndex(index, nil, nil)
	if err != nil {
		status, err = prettyPrintCorruptedIndexFix(index)
		if err != nil {
			return nil
		}
	}

	libraryOK := make([]string, 0, len(args))
	libraryFails := make(map[string]string, len(args))

	for _, libraryName := range args {
		library := status.Libraries[libraryName]
		if library != nil {
			//found
			_, err = libraries.DownloadAndCache(library)
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
		prettyPrintDownload(libraryOK, libraryFails)
	} else {
		for _, library := range libraryOK {
			fmt.Printf("%s - Downloaded\n", library)
		}
		for library, failure := range libraryFails {
			fmt.Printf("%s - Error : %s\n", library, failure)
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
		err = prettyPrintDownloadFileIndex()
		if err != nil {
			return nil
		}
	}

	status, err := libraries.CreateStatusContextFromIndex(index, nil, nil)
	if err != nil {
		status, err = prettyPrintCorruptedIndexFix(index)
		if err != nil {
			return nil
		}
	}

	libraryOK := make([]string, 0, len(args))
	libraryFails := make(map[string]string, len(args))

	for _, libraryName := range args {
		library := status.Libraries[libraryName]
		if library != nil {
			//found
			err = libraries.DownloadAndInstall(library)
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
		prettyPrintInstall(libraryOK, libraryFails)
	} else {
		for _, library := range libraryOK {
			fmt.Printf("%s - Installed\n", library)
		}
		for library, failure := range libraryFails {
			fmt.Printf("%s - Error : %s\n", library, failure)
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
					strings.Replace(fileName, "_", " ", 0)
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

	if len(libraryFails) > 0 {
		fmt.Println("The following libraries were succesfully uninstalled:")
		fmt.Println(strings.Join(libraryOK, " "))
		fmt.Println("However, the uninstall process failed on the following libraries:")
		fmt.Println(strings.Join(libraryFails, " "))
	} else {
		fmt.Println("All libraries successfully uninstalled")
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

	status, err := libraries.CreateStatusContextFromIndex(index, nil, nil)
	if err != nil {
		if GlobalFlags.Verbose > 0 {
			fmt.Println("Cannot parse index file, it may be corrupted. downloading from downloads.arduino.cc")
		}

		err = prettyPrintDownloadFileIndex()
		if err != nil {
			return nil
		}

		//after download, I retry.
		status, err = libraries.CreateStatusContextFromIndex(index, nil, nil)
		if err != nil {
			if GlobalFlags.Verbose > 0 {
				fmt.Println("ERROR")
			}
			fmt.Println("Cannot parse downloaded index file")
			return nil
		}
		if GlobalFlags.Verbose > 0 {
			fmt.Println("OK")
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

func prettyPrintInstall(libraryOK []string, libraryFails map[string]string) {
	if len(libraryFails) > 0 {
		if len(libraryOK) > 0 {
			fmt.Println("The following libraries were succesfully installed:")
			fmt.Println(strings.Join(libraryOK, " "))
			fmt.Print("However, t")
		} else { //UGLYYYY but it works
			fmt.Print("T")
		}
		fmt.Println("he installation process failed on the following libraries:")
		for library, failure := range libraryFails {
			fmt.Printf("%s - %s\n", library, failure)
		}
	} else {
		fmt.Println("All libraries successfully installed")
	}
}

//TODO: remove copypasting from prettyPrintInstall and merge them in a single function
func prettyPrintDownload(libraryOK []string, libraryFails map[string]string) {
	if len(libraryFails) > 0 {
		if len(libraryOK) > 0 {
			fmt.Println("The following libraries were succesfully downloaded:")
			fmt.Println(strings.Join(libraryOK, " "))
			fmt.Print("However, t")
		} else { //UGLYYYY but it works
			fmt.Print("T")
		}
		fmt.Println("he download of the following libraries failed:")
		for library, failure := range libraryFails {
			fmt.Printf("%s - %s\n", library, failure)
		}
	} else {
		fmt.Println("All libraries successfully downloaded")
	}
}

func prettyPrintCorruptedIndexFix(index *libraries.Index) (*libraries.StatusContext, error) {
	if GlobalFlags.Verbose > 0 {
		fmt.Println("Cannot parse index file, it may be corrupted.")
	}

	err := prettyPrintDownloadFileIndex()
	if err != nil {
		return nil, err
	}

	return prettyIndexParse(index)
}

func prettyIndexParse(index *libraries.Index) (*libraries.StatusContext, error) {
	if GlobalFlags.Verbose > 0 {
		fmt.Print("Parsing downloaded index file ... ")
	}

	//after download, I retry.
	status, err := libraries.CreateStatusContextFromIndex(index, nil, nil)
	if err != nil {
		if GlobalFlags.Verbose > 0 {
			fmt.Println("ERROR")
		}
		fmt.Println("Cannot parse index file")
		return nil, err
	}
	if GlobalFlags.Verbose > 0 {
		fmt.Println("OK")
	}

	return status, nil
}
