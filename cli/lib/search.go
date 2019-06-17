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
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	semver "go.bug.st/relaxed-semver"
)

func initSearchCommand() *cobra.Command {
	searchCommand := &cobra.Command{
		Use:     "search [LIBRARY_NAME]",
		Short:   "Searchs for one or more libraries data.",
		Long:    "Search for one or more libraries data (case insensitive search).",
		Example: "  " + cli.VersionInfo.Application + " lib search audio",
		Args:    cobra.ArbitraryArgs,
		Run:     runSearchCommand,
	}
	searchCommand.Flags().BoolVar(&searchFlags.namesOnly, "names", false, "Show library names only.")
	return searchCommand
}

var searchFlags struct {
	namesOnly bool // if true outputs lib names only.
}

func runSearchCommand(cmd *cobra.Command, args []string) {
	instance := cli.CreateInstaceIgnorePlatformIndexErrors()
	logrus.Info("Executing `arduino lib search`")
	searchResp, err := lib.LibrarySearch(context.Background(), &rpc.LibrarySearchReq{
		Instance: instance,
		Query:    (strings.Join(args, " ")),
	})
	if err != nil {
		formatter.PrintError(err, "Error saerching for Library")
		os.Exit(cli.ErrGeneric)
	}

	if cli.OutputJSONOrElse(searchResp) {
		results := searchResp.GetLibraries()
		sort.Slice(results, func(i, j int) bool {
			return results[i].Name < results[j].Name
		})
		if searchFlags.namesOnly {
			for _, result := range results {
				fmt.Println(result.Name)
			}
		} else {
			if len(results) > 0 {
				for _, result := range results {
					outputSearchedLibrary(result)
				}
			} else {
				formatter.Print("No libraries matching your search.")
			}
		}
	}
	logrus.Info("Done")
}

func outputSearchedLibrary(lsr *rpc.SearchedLibrary) {
	fmt.Printf("Name: \"%s\"\n", lsr.Name)
	fmt.Printf("  Author: %s\n", lsr.GetLatest().Author)
	fmt.Printf("  Maintainer: %s\n", lsr.GetLatest().Maintainer)
	fmt.Printf("  Sentence: %s\n", lsr.GetLatest().Sentence)
	fmt.Printf("  Paragraph: %s\n", lsr.GetLatest().Paragraph)
	fmt.Printf("  Website: %s\n", lsr.GetLatest().Website)
	fmt.Printf("  Category: %s\n", lsr.GetLatest().Category)
	fmt.Printf("  Architecture: %s\n", strings.Join(lsr.GetLatest().Architectures, ", "))
	fmt.Printf("  Types: %s\n", strings.Join(lsr.GetLatest().Types, ", "))
	fmt.Printf("  Versions: %s\n", strings.Replace(fmt.Sprint(versionsFromSearchedLibrary(lsr)), " ", ", ", -1))
}

func versionsFromSearchedLibrary(library *rpc.SearchedLibrary) []*semver.Version {
	res := []*semver.Version{}
	for str := range library.Releases {
		if v, err := semver.Parse(str); err == nil {
			res = append(res, v)
		}
	}
	sort.Sort(semver.List(res))
	return res
}
