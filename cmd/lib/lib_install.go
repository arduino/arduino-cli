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

package libCmd

import (
	"fmt"
	"strings"

	"github.com/bcmi-labs/arduino-cli/libraries"
	"github.com/spf13/cobra"
)

// installCmd represents the lib install command.
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs one of more specified libraries into the system.",
	Long:  `Installs one or more specified libraries into the system.`,
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
		if GlobalFlags.Verbose > 0 {
			fmt.Println("Cannot parse index file, it may be corrupted. downloading from downloads.arduino.cc")
		}

		err = libraries.DownloadLibrariesFile()
		if err != nil {
			if GlobalFlags.Verbose > 0 {
				fmt.Println("ERROR")
			}
			fmt.Println("Cannot download index file, please check your network connection.")
			return nil
		}
		if GlobalFlags.Verbose > 0 {
			fmt.Println("OK")
		}

		if GlobalFlags.Verbose > 0 {
			fmt.Print("Parsing downloaded index file ... ")
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
		prettyPrint(libraryOK, libraryFails)
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

func prettyPrintInstall(libraryOK []string, libraryFails map[string]string) {
	if len(libraryFails) > 0 {
		fmt.Println("The following libraries were succesfully installed:")
		fmt.Println(strings.Join(libraryOK, " "))
		fmt.Println("However, the installation process failed on the following libraries:")
		for library, failure := range libraryFails {
			fmt.Printf("%s - %s\n", library, failure)
		}
	} else {
		fmt.Println("All libraries successfully installed")
	}
}
