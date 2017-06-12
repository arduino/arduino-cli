// Copyright © 2017 Alessandro Sanino <saninoale@gmail.com>
// {{.copyright}}
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"fmt"

	"github.com/bcmi-labs/arduino-cli/libraries"
	"github.com/spf13/cobra"
)

// installCmd represents the lib install command.
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs a specified library into the system.",
	Long:  `Installs a specified library into the system.`,
	RunE:  executeInstallCommand,
}

func init() {
	LibRoot.AddCommand(installCmd)
}

func executeInstallCommand(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("No library specified for install command")
	}

	index, err := libraries.LoadLibrariesIndex()
	if err != nil {
		fmt.Print("Cannot find index file, downloading from downloads.arduino.cc ... ")
		err = libraries.DownloadLibrariesFile()
		if err != nil {
			fmt.Println("ERROR")
			fmt.Println("Cannot download index file, please check your network connection.")
			return nil
		}
		fmt.Println("OK")
	}

	status, err := libraries.CreateStatusContextFromIndex(index, nil, nil)
	if err != nil {
		fmt.Println("Cannot parse index file, it may be corrupted. Try run arduino lib list update to update the index.")
	}

	libraryFails := make(map[string]string, len(args))

	for _, libraryName := range args {
		library := status.Libraries[libraryName]
		if library != nil {
			//found
			err = libraries.DownloadAndInstall(library)
			if err != nil {
				libraryFails[libraryName] = err.Error()
			}
		} else {
			libraryFails[libraryName] = "This library is not in library index"
		}
	}

	if len(libraryFails) > 0 {
		fmt.Println("Installation encountered the following error(s):")
		for library, failure := range libraryFails {
			fmt.Printf("%s - %s\n", library, failure)
		}
	} else {
		fmt.Println("Install Success")
	}

	return nil
}
