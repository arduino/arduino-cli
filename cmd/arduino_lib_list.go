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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"os"

	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/spf13/cobra"
	"github.com/zieckey/goini"
)

// arduinoLibListCmd represents the list libraries command.
var arduinoLibListCmd = &cobra.Command{
	Use:   "list",
	Short: "Shows a list of all installed libraries",
	Long: `Shows a list of all installed libraries.
Can be used with -v (or --verbose) flag (up to 2 times) to have longer output.`,
	Run: executeListCommand,
}

// arduinoLibListUpdateCmd represents the lib list update command
var arduinoLibListUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates the library index to latest version",
	Long:  `Updates the library index to latest version from downloads.arduino.cc repository.`,
	Run:   execUpdateListIndex,
}

func init() {
	arduinoLibCmd.AddCommand(arduinoLibListCmd)
	arduinoLibListCmd.AddCommand(arduinoLibListUpdateCmd)
}

func executeListCommand(command *cobra.Command, args []string) {
	libHome, err := common.GetDefaultLibFolder()
	if err != nil {
		fmt.Println("Cannot get libraries folder")
		return
	}

	//prettyPrintStatus(status)
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
	prettyPrintDownloadFileIndex()
}
