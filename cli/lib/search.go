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
	"context"
	"fmt"
	"os"
	"sort"

	"strings"

	"github.com/arduino/arduino-cli/cli"
	"github.com/arduino/arduino-cli/commands/lib"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/rpc"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	semver "go.bug.st/relaxed-semver"
)

func initSearchCommand() *cobra.Command {
	searchCommand := &cobra.Command{
		Use:     "search [LIBRARY_NAME]",
		Short:   "Searchs for one or more libraries data.",
		Long:    "Search for one or more libraries data (case insensitive search).",
		Example: "  " + cli.AppName + " lib search audio",
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
	instance := cli.CreateInstance()
	logrus.Info("Executing `arduino lib search`")
	//arguments :=
	searchResp, err := lib.LibrarySearch(context.Background(), &rpc.LibrarySearchReq{
		Instance: instance,
		Names:    searchFlags.names,
		Query:    (strings.Join(args, " ")),
	})
	if err != nil {
		formatter.PrintError(err, "Error saerching for Library")
		os.Exit(cli.ErrGeneric)
	}

	results := searchResp.GetSearchOutput()
	if cli.OutputJSONOrElse(results) {
		if len(results) > 0 {
			for _, out := range results {
				fmt.Println(SearchOutputToString(out, searchFlags.names))
			}
		} else {
			formatter.Print("No libraries matching your search.")
		}
	}
	logrus.Info("Done")
}

func SearchOutputToString(lsr *rpc.SearchLibraryOutput, names bool) string {
	ret := ""

	ret += fmt.Sprintf("Name: \"%s\"\n", lsr.Name)
	if !names {
		ret += fmt.Sprintln("  Author: ", lsr.GetLatest().Author) +
			fmt.Sprintln("  Maintainer: ", lsr.GetLatest().Maintainer) +
			fmt.Sprintln("  Sentence: ", lsr.GetLatest().Sentence) +
			fmt.Sprintln("  Paragraph: ", lsr.GetLatest().Paragraph) +
			fmt.Sprintln("  Website: ", lsr.GetLatest().Website) +
			fmt.Sprintln("  Category: ", lsr.GetLatest().Category) +
			fmt.Sprintln("  Architecture: ", strings.Join(lsr.GetLatest().Architectures, ", ")) +
			fmt.Sprintln("  Types: ", strings.Join(lsr.GetLatest().Types, ", ")) +
			fmt.Sprintln("  Versions: ", strings.Replace(fmt.Sprint(Versions(lsr.GetLatest(), lsr)), " ", ", ", -1))
	}
	return strings.TrimSpace(ret)
}

func Versions(l *rpc.LibraryRelease, library *rpc.SearchLibraryOutput) []*semver.Version {
	res := []*semver.Version{}
	for str, _ := range library.Releases {
		v, err := semver.Parse(str)
		if err == nil {
			res = append(res, v)
		}
	}
	sort.Sort(semver.List(res))
	return res
}
