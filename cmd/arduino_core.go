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
	"os"
	"path/filepath"
	"strings"

	"github.com/bcmi-labs/arduino-cli/cmd/formatter"
	"github.com/bcmi-labs/arduino-cli/cmd/output"
	"github.com/bcmi-labs/arduino-cli/cmd/pretty_print"
	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/spf13/cobra"
	"github.com/zieckey/goini"
)

var arduinoCoreCmd = &cobra.Command{
	Use:   "core",
	Short: "Arduino Core operations",
	Long:  `Arduino Core operations`,
	Run:   executeCoreCommand,
}

var arduinoCoreListCmd = &cobra.Command{
	Use:   "list",
	Short: "Shows the list of installed cores",
	Long: `Shows the list of installed cores. 
With -v tag (up to 2 times) can provide more verbose output.`,
	Run: executeCoreListCommand,
}

var arduinoCoreFlags struct {
	updateIndex bool
}

func init() {
	arduinoCmd.AddCommand(arduinoCoreCmd)
	arduinoCoreCmd.AddCommand(arduinoCoreListCmd)

	arduinoCoreCmd.Flags().BoolVar(&arduinoCoreFlags.updateIndex, "update-index", false, "Updates the index of cores to the latest version")
}

func executeCoreCommand(cmd *cobra.Command, args []string) {
	if arduinoCoreFlags.updateIndex {
		common.ExecUpdateIndex(prettyPrints.DownloadCoreFileIndex(), GlobalFlags.Verbose)
	} else {
		cmd.Help()
	}
}

func executeCoreListCommand(cmd *cobra.Command, args []string) {
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
