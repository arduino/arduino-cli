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
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package lib

import (
	"fmt"
	"os"
	"strings"

	"github.com/bcmi-labs/arduino-cli/arduino/libraries"
	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	command.AddCommand(searchCommand)
	searchCommand.Flags().BoolVar(&searchFlags.names, "names", false, "Show library names only.")
}

var searchFlags struct {
	names bool // if true outputs lib names only.
}

var searchCommand = &cobra.Command{
	Use:   "search [LIBRARY_NAME]",
	Short: "Searchs for one or more libraries data.",
	Long:  "Search for one or more libraries data (case insensitive search).",
	Example: "" +
		"arduino lib search You # to show all libraries containing \"You\" in their name (case insensitive).\n" +
		"YoumadeIt\n" +
		"YoutubeApi",
	Args: cobra.ArbitraryArgs,
	Run:  runSearchCommand,
}

func runSearchCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino lib search`")
	query := strings.ToLower(strings.Join(args, " "))

	logrus.Info("Getting libraries status context")
	status, err := getLibStatusContext()
	if err != nil {
		formatter.PrintError(err, "Error while parsing installed libraries.")
		os.Exit(commands.ErrCoreConfig)
	}

	res := output.LibSearchResults{
		Libraries: []*libraries.Library{},
	}
	for _, lib := range status.Libraries {
		if strings.Contains(strings.ToLower(lib.Name), query) {
			res.Libraries = append(res.Libraries, lib)
		}
	}

	if len(res.Libraries) == 0 {
		formatter.Print(fmt.Sprintf("No library found matching `%s` search query", query))
	} else {
		formatter.Print(res)
	}
	logrus.Info("Done")
}
