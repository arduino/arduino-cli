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

	"github.com/bcmi-labs/arduino-cli/libraries"
	"github.com/spf13/cobra"
)

// arduinoLibListCmd represents the list libraries command.
var arduinoLibListCmd = &cobra.Command{
	Use:   "list",
	Short: "Shows a list of all libraries from arduino repository.",
	Long: `Shows a list of all libraries from arduino repository.
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
	//fmt.Println("libs list:", args)
	//fmt.Println("long =", libListCmdFlags.Long)
	libFile, _ := libraries.IndexPath()

	//If it doesn't exist download it
	if _, err := os.Stat(libFile); os.IsNotExist(err) {
		if GlobalFlags.Verbose > 0 {
			fmt.Println("Index file not found ... ")
		}

		err = prettyPrintDownloadFileIndex()
		if err != nil {
			return
		}
	}

	//If it exists but it is corrupt replace it from arduino repository.
	index, err := libraries.LoadLibrariesIndex()
	if err != nil {
		err = prettyPrintDownloadFileIndex()
		if err != nil {
			return
		}

		index, err = libraries.LoadLibrariesIndex()
		if err != nil {
			fmt.Println("Cannot parse index file")
			return
		}
	}

	status, err := libraries.CreateStatusContextFromIndex(index, nil, nil)
	if err != nil {
		status, err = prettyPrintCorruptedIndexFix(index)
		if err != nil {
			return
		}
	}

	prettyPrintStatus(status)
}

func prettyPrintStatus(status *libraries.StatusContext) {
	//Pretty print libraries from index.
	for _, name := range status.Names() {
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

func execUpdateListIndex(cmd *cobra.Command, args []string) {
	prettyPrintDownloadFileIndex()
}

func prettyPrintDownloadFileIndex() error {
	if GlobalFlags.Verbose > 0 {
		fmt.Print("Downloading a new index file from download.arduino.cc ... ")
	}

	err := libraries.DownloadLibrariesFile()
	if err != nil {
		if GlobalFlags.Verbose > 0 {
			fmt.Println("ERROR")
		}
		fmt.Println("Cannot download index file, check your network connection.")
		return err
	}

	if GlobalFlags.Verbose > 0 {
		fmt.Println("OK")
	}

	return nil
}
