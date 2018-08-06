/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package lib

import (
	"fmt"
	"strings"

	"github.com/bcmi-labs/arduino-cli/commands"

	"github.com/bcmi-labs/arduino-cli/arduino/libraries/librariesindex"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initSearchCommand() *cobra.Command {
	searchCommand := &cobra.Command{
		Use:     "search [LIBRARY_NAME]",
		Short:   "Searchs for one or more libraries data.",
		Long:    "Search for one or more libraries data (case insensitive search).",
		Example: "arduino lib search audio",
		Args:    cobra.ArbitraryArgs,
		Run:     runSearchCommand,
	}
	searchCommand.Flags().BoolVar(&searchFlags.names, "names", false, "Show library names only.")
	return searchCommand
}

var searchFlags struct {
	names bool // if true outputs lib names only.
}

func runSearchCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino lib search`")
	query := strings.ToLower(strings.Join(args, " "))

	lm := commands.InitLibraryManager(nil)

	res := output.LibSearchResults{
		Libraries: []*librariesindex.Library{},
	}
	for _, lib := range lm.Index.Libraries {
		if strings.Contains(strings.ToLower(lib.Name), query) {
			res.Libraries = append(res.Libraries, lib)
		}
	}

	if searchFlags.names {
		for _, lib := range res.Libraries {
			formatter.Print(lib.Name)
		}
	} else {
		if len(res.Libraries) == 0 {
			formatter.Print(fmt.Sprintf("No library found matching `%s` search query", query))
		} else {
			formatter.Print(res)
		}
	}
	logrus.Info("Done")
}
