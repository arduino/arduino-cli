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
	"os"
	"path/filepath"
	"strings"

	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/bcmi-labs/arduino-cli/libraries"
	"github.com/spf13/cobra"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstalls one or more libraries",
	Long:  `Uninstalls one or more libraries`,
	RunE:  executeUninstallCommand,
}

func init() {
	LibRoot.AddCommand(uninstallCmd)
}

func executeUninstallCommand(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("No library specified for install command")
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

	dirFiles, err := dir.Readdirnames(0)
	if err != nil {
		fmt.Println("Cannot read into libraries folder")
		return nil
	}

	for _, fileName := range dirFiles {
		for _, library := range args {
			if strings.Contains(fileName, library) {
				//found
				//TODO : remove only one version
				err = libraries.Uninstall(filepath.Join(libFolder, fileName))
				if err != nil {
					libraryFails = append(libraryFails, library)
				} else {
					libraryOK = append(libraryOK, library)
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
