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
	"path/filepath"

	"github.com/arduino/arduino-cli/libraries"
	"github.com/spf13/cobra"
)

// libListCmd represents the list libraries command.
var libListCmd = &cobra.Command{
	Use:   "list",
	Short: "Shows a list of all libraries from arduino repository.",
	Long: `Shows a list of all libraries from arduino repository.
Can be used with -v (or --verbose) flag (up to 2 times) to have longer output.`,
	Run: executeListCommand,
}

var libListCmdFlags struct {
	Verbose int
}

func init() {
	LibRoot.AddCommand(libListCmd)
}

func executeListCommand(command *cobra.Command, args []string) {
	//fmt.Println("libs list:", args)
	//fmt.Println("long =", libListCmdFlags.Long)
	baseFolder, err := GetDefaultArduinoFolder()
	if err != nil {
		fmt.Printf("Could not determine data folder: %s", err)
	}

	libFile := filepath.Join(baseFolder, "library_index.json")
	index, err := libraries.LoadLibrariesIndex(libFile)
	if err != nil {
		fmt.Printf("Could not read library index: %s", err)
	}

	//fmt.Printf("libFile = %s\n", libFile)
	//fmt.Printf("index = %v\n", index)

	libraries, err := libraries.CreateStatusContextFromIndex(index, nil, nil)
	if err != nil {
		fmt.Printf("Could not synchronize library status: %s", err)
	}

	for _, name := range libraries.Names() {
		if GlobalFlags.Verbose > 0 {
			lib := libraries.Libraries[name]
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
