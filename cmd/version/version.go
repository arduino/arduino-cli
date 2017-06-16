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

package versionCmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	arduino "github.com/bcmi-labs/arduino-cli/cmd"
	lib "github.com/bcmi-labs/arduino-cli/cmd/lib"
)

// cliVersionCmd represents the version command.
var cliVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows version Number of arduino",
	Long:  `Shows version Number of arduino which is installed on your system.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("arduino V. %s\n", arduino.ArduinoVersion)
		if arduino.GlobalFlags.Verbose > 0 {
			fmt.Printf("arduino V. %s\n", lib.LibVersion)
		}
	},
}

// TODO : maybe it is possible to autogenerate versions from ancestors, as I wrote in this function.
// suggestions : save in a map[string]*cobra.Command variable all info about templates which vary only by command name and
// version number.
func ancestorsBreadcrumb(cmd *cobra.Command) string {
	ancestors := make([]string, 2, 2)
	cmd.VisitParents(func(ancestor *cobra.Command) {
		ancestors = append(ancestors, ancestor.Use)
	})
	//fmt.Println(ancestors)
	return strings.Trim(strings.Join(ancestors, " "), " ")
}

func init() {
	arduino.RootCmd.AddCommand(cliVersionCmd)
}
