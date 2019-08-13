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

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/lib"
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
		Example: "  " + os.Args[0] + " lib search audio",
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
	instance := instance.CreateInstaceIgnorePlatformIndexErrors()
	logrus.Info("Executing `arduino lib search`")
	searchResp, err := lib.LibrarySearch(context.Background(), &rpc.LibrarySearchReq{
		Instance: instance,
		Query:    (strings.Join(args, " ")),
	})
	if err != nil {
		feedback.Errorf("Error saerching for Library: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	if globals.OutputFormat == "json" {
		feedback.PrintJSON(searchResp)
	} else {
		// get a sorted slice of results
		results := searchResp.GetLibraries()
		sort.Slice(results, func(i, j int) bool {
			return results[i].Name < results[j].Name
		})

		// print all the things
		outputSearchedLibrary(results, searchFlags.namesOnly)
	}

	logrus.Info("Done")
}

func outputSearchedLibrary(results []*rpc.SearchedLibrary, namesOnly bool) {
	if len(results) == 0 {
		feedback.Print("No libraries matching your search.")
		return
	}

	for _, lsr := range results {
		feedback.Printf("Name: '%s'", lsr.Name)
		if namesOnly {
			continue
		}

		feedback.Printf("  Author: %s", lsr.GetLatest().Author)
		feedback.Printf("  Maintainer: %s", lsr.GetLatest().Maintainer)
		feedback.Printf("  Sentence: %s", lsr.GetLatest().Sentence)
		feedback.Printf("  Paragraph: %s", lsr.GetLatest().Paragraph)
		feedback.Printf("  Website: %s", lsr.GetLatest().Website)
		feedback.Printf("  Category: %s", lsr.GetLatest().Category)
		feedback.Printf("  Architecture: %s", strings.Join(lsr.GetLatest().Architectures, ", "))
		feedback.Printf("  Types: %s", strings.Join(lsr.GetLatest().Types, ", "))
		feedback.Printf("  Versions: %s", strings.Replace(fmt.Sprint(versionsFromSearchedLibrary(lsr)), " ", ", ", -1))
	}
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
