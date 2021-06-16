// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package lib

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/lib"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	semver "go.bug.st/relaxed-semver"
)

func initSearchCommand() *cobra.Command {
	searchCommand := &cobra.Command{
		Use:     "search [LIBRARY_NAME]",
		Short:   "Searches for one or more libraries data.",
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
	inst, status := instance.Create()
	if status != nil {
		feedback.Errorf("Error creating instance: %v", status)
		os.Exit(errorcodes.ErrGeneric)
	}

	err := commands.UpdateLibrariesIndex(context.Background(), &rpc.UpdateLibrariesIndexRequest{
		Instance: inst,
	}, output.ProgressBar())
	if err != nil {
		feedback.Errorf("Error updating library index: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	for _, err := range instance.Init(inst) {
		feedback.Errorf("Error initializing instance: %v", err)
	}

	logrus.Info("Executing `arduino lib search`")
	searchResp, err := lib.LibrarySearch(context.Background(), &rpc.LibrarySearchRequest{
		Instance: inst,
		Query:    (strings.Join(args, " ")),
	})
	if err != nil {
		feedback.Errorf("Error searching for Library: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	feedback.PrintResult(result{
		results:   searchResp,
		namesOnly: searchFlags.namesOnly,
	})

	logrus.Info("Done")
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type result struct {
	results   *rpc.LibrarySearchResponse
	namesOnly bool
}

func (res result) Data() interface{} {
	if res.namesOnly {
		type LibName struct {
			Name string `json:"name,required"`
		}

		type NamesOnly struct {
			Libraries []LibName `json:"libraries,required"`
		}

		names := []LibName{}
		results := res.results.GetLibraries()
		for _, lib := range results {
			names = append(names, LibName{lib.Name})
		}

		return NamesOnly{
			names,
		}
	}

	return res.results
}

func (res result) String() string {
	results := res.results.GetLibraries()
	if len(results) == 0 {
		return "No libraries matching your search."
	}

	// get a sorted slice of results
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	var out strings.Builder

	if res.results.GetStatus() == rpc.LibrarySearchStatus_LIBRARY_SEARCH_STATUS_FAILED {
		out.WriteString("No libraries matching your search.\nDid you mean...\n")
	}

	for _, lib := range results {
		if res.results.GetStatus() == rpc.LibrarySearchStatus_LIBRARY_SEARCH_STATUS_SUCCESS {
			out.WriteString(fmt.Sprintf("Name: \"%s\"\n", lib.Name))
			if res.namesOnly {
				continue
			}
		} else {
			out.WriteString(fmt.Sprintf("%s\n", lib.Name))
			continue
		}

		latest := lib.GetLatest()

		deps := []string{}
		for _, dep := range latest.GetDependencies() {
			if dep.GetVersionConstraint() == "" {
				deps = append(deps, dep.GetName())
			} else {
				deps = append(deps, dep.GetName()+" ("+dep.GetVersionConstraint()+")")
			}
		}

		out.WriteString(fmt.Sprintf("  Author: %s\n", latest.Author))
		out.WriteString(fmt.Sprintf("  Maintainer: %s\n", latest.Maintainer))
		out.WriteString(fmt.Sprintf("  Sentence: %s\n", latest.Sentence))
		out.WriteString(fmt.Sprintf("  Paragraph: %s\n", latest.Paragraph))
		out.WriteString(fmt.Sprintf("  Website: %s\n", latest.Website))
		if latest.License != "" {
			out.WriteString(fmt.Sprintf("  License: %s\n", latest.License))
		}
		out.WriteString(fmt.Sprintf("  Category: %s\n", latest.Category))
		out.WriteString(fmt.Sprintf("  Architecture: %s\n", strings.Join(latest.Architectures, ", ")))
		out.WriteString(fmt.Sprintf("  Types: %s\n", strings.Join(latest.Types, ", ")))
		out.WriteString(fmt.Sprintf("  Versions: %s\n", strings.Replace(fmt.Sprint(versionsFromSearchedLibrary(lib)), " ", ", ", -1)))
		if len(latest.ProvidesIncludes) > 0 {
			out.WriteString(fmt.Sprintf("  Provides includes: %s\n", strings.Join(latest.ProvidesIncludes, ", ")))
		}
		if len(latest.Dependencies) > 0 {
			out.WriteString(fmt.Sprintf("  Dependencies: %s\n", strings.Join(deps, ", ")))
		}
	}

	return fmt.Sprintf("%s", out.String())
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
