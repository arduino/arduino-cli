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
	"strings"
	"time"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/lib"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/feedback/result"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initSearchCommand() *cobra.Command {
	var namesOnly bool
	var omitReleasesDetails bool
	searchCommand := &cobra.Command{
		Use:   fmt.Sprintf("search [%s ...]", tr("SEARCH_TERM")),
		Short: tr("Searches for one or more libraries matching a query."),
		Long: tr(`Search for libraries matching zero or more search terms.

All searches are performed in a case-insensitive fashion. Queries containing
multiple search terms will return only libraries that match all of the terms.

Search terms that do not match the QV syntax described below are basic search
terms, and will match libraries that include the term anywhere in any of the
following fields:
 - Author
 - Name
 - Paragraph
 - Provides
 - Sentence

A special syntax, called qualifier-value (QV), indicates that a search term
should be compared against only one field of each library index entry. This
syntax uses the name of an index field (case-insensitive), an equals sign (=)
or a colon (:), and a value, e.g. 'name=ArduinoJson' or 'provides:tinyusb.h'.

QV search terms that use a colon separator will match all libraries with the
value anywhere in the named field, and QV search terms that use an equals
separator will match only libraries with exactly the provided value in the
named field.

QV search terms can include embedded spaces using double-quote (") characters
around the value or the entire term, e.g. 'category="Data Processing"' and
'"category=Data Processing"' are equivalent. A QV term can include a literal
double-quote character by preceding it with a backslash (\) character.

NOTE: QV search terms using double-quote or backslash characters that are
passed as command-line arguments may require quoting or escaping to prevent
the shell from interpreting those characters.

In addition to the fields listed above, QV terms can use these qualifiers:
 - Architectures
 - Category
 - Dependencies
 - License
 - Maintainer
 - Types
 - Version
 - Website
		`),
		Example: "  " + os.Args[0] + " lib search audio                               # " + tr("basic search for \"audio\"") + "\n" +
			"  " + os.Args[0] + " lib search name:buzzer                         # " + tr("libraries with \"buzzer\" in the Name field") + "\n" +
			"  " + os.Args[0] + " lib search name=pcf8523                        # " + tr("libraries with a Name exactly matching \"pcf8523\"") + "\n" +
			"  " + os.Args[0] + " lib search \"author:\\\"Daniel Garcia\\\"\"          # " + tr("libraries authored by Daniel Garcia") + "\n" +
			"  " + os.Args[0] + " lib search author=Adafruit name:gfx            # " + tr("libraries authored only by Adafruit with \"gfx\" in their Name") + "\n" +
			"  " + os.Args[0] + " lib search esp32 display maintainer=espressif  # " + tr("basic search for \"esp32\" and \"display\" limited to official Maintainer") + "\n" +
			"  " + os.Args[0] + " lib search dependencies:IRremote               # " + tr("libraries that depend on at least \"IRremote\"") + "\n" +
			"  " + os.Args[0] + " lib search dependencies=IRremote               # " + tr("libraries that depend only on \"IRremote\"") + "\n",
		Args: cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runSearchCommand(args, namesOnly, omitReleasesDetails)
		},
	}
	searchCommand.Flags().BoolVar(&namesOnly, "names", false, tr("Show library names only."))
	searchCommand.Flags().BoolVar(&omitReleasesDetails, "omit-releases-details", false, tr("Omit library details far all versions except the latest (produce a more compact JSON output)."))
	return searchCommand
}

// indexUpdateInterval specifies the time threshold over which indexes are updated
const indexUpdateInterval = 60 * time.Minute

func runSearchCommand(args []string, namesOnly bool, omitReleasesDetails bool) {
	inst := instance.CreateAndInit()

	logrus.Info("Executing `arduino-cli lib search`")

	if indexNeedsUpdating(indexUpdateInterval) {
		if err := commands.UpdateLibrariesIndex(
			context.Background(),
			&rpc.UpdateLibrariesIndexRequest{Instance: inst},
			feedback.ProgressBar(),
		); err != nil {
			feedback.Fatal(tr("Error updating library index: %v", err), feedback.ErrGeneric)
		}
		instance.Init(inst)
	}

	searchResp, err := lib.LibrarySearch(context.Background(), &rpc.LibrarySearchRequest{
		Instance:            inst,
		SearchArgs:          strings.Join(args, " "),
		OmitReleasesDetails: omitReleasesDetails,
	})
	if err != nil {
		feedback.Fatal(tr("Error searching for Libraries: %v", err), feedback.ErrGeneric)
	}

	feedback.PrintResult(librarySearchResult{
		results:   result.NewLibrarySearchResponse(searchResp),
		namesOnly: namesOnly,
	})

	logrus.Info("Done")
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type librarySearchResult struct {
	results   *result.LibrarySearchResponse
	namesOnly bool
}

func (res librarySearchResult) Data() interface{} {
	if res.namesOnly {
		type LibName struct {
			Name string `json:"name"`
		}

		type NamesOnly struct {
			Libraries []LibName `json:"libraries"`
		}

		names := []LibName{}
		for _, lib := range res.results.Libraries {
			if lib == nil {
				continue
			}
			names = append(names, LibName{lib.Name})
		}

		return NamesOnly{names}
	}

	return res.results
}

func (res librarySearchResult) String() string {
	results := res.results.Libraries
	if len(results) == 0 {
		return tr("No libraries matching your search.")
	}

	var out strings.Builder

	if string(res.results.Status) == rpc.LibrarySearchStatus_name[int32(rpc.LibrarySearchStatus_LIBRARY_SEARCH_STATUS_FAILED)] {
		out.WriteString(tr("No libraries matching your search.\nDid you mean...\n"))
	}

	for _, lib := range results {
		if lib == nil {
			continue
		}
		if string(res.results.Status) == rpc.LibrarySearchStatus_name[int32(rpc.LibrarySearchStatus_LIBRARY_SEARCH_STATUS_SUCCESS)] {
			out.WriteString(tr(`Name: "%s"`, lib.Name) + "\n")
			if res.namesOnly {
				continue
			}
		} else {
			out.WriteString(fmt.Sprintf("%s\n", lib.Name))
			continue
		}

		latest := lib.Latest

		deps := []string{}
		for _, dep := range latest.Dependencies {
			if dep == nil {
				continue
			}
			if dep.VersionConstraint == "" {
				deps = append(deps, dep.Name)
			} else {
				deps = append(deps, dep.Name+" ("+dep.VersionConstraint+")")
			}
		}

		out.WriteString(fmt.Sprintf("  "+tr("Author: %s")+"\n", latest.Author))
		out.WriteString(fmt.Sprintf("  "+tr("Maintainer: %s")+"\n", latest.Maintainer))
		out.WriteString(fmt.Sprintf("  "+tr("Sentence: %s")+"\n", latest.Sentence))
		out.WriteString(fmt.Sprintf("  "+tr("Paragraph: %s")+"\n", latest.Paragraph))
		out.WriteString(fmt.Sprintf("  "+tr("Website: %s")+"\n", latest.Website))
		if latest.License != "" {
			out.WriteString(fmt.Sprintf("  "+tr("License: %s")+"\n", latest.License))
		}
		out.WriteString(fmt.Sprintf("  "+tr("Category: %s")+"\n", latest.Category))
		out.WriteString(fmt.Sprintf("  "+tr("Architecture: %s")+"\n", strings.Join(latest.Architectures, ", ")))
		out.WriteString(fmt.Sprintf("  "+tr("Types: %s")+"\n", strings.Join(latest.Types, ", ")))
		out.WriteString(fmt.Sprintf("  "+tr("Versions: %s")+"\n", strings.ReplaceAll(fmt.Sprint(lib.AvailableVersions), " ", ", ")))
		if len(latest.ProvidesIncludes) > 0 {
			out.WriteString(fmt.Sprintf("  "+tr("Provides includes: %s")+"\n", strings.Join(latest.ProvidesIncludes, ", ")))
		}
		if len(latest.Dependencies) > 0 {
			out.WriteString(fmt.Sprintf("  "+tr("Dependencies: %s")+"\n", strings.Join(deps, ", ")))
		}
	}

	return out.String()
}

// indexNeedsUpdating returns whether library_index.json needs updating
func indexNeedsUpdating(timeout time.Duration) bool {
	// Library index path is constant (relative to the data directory).
	// It does not depend on board manager URLs or any other configuration.
	dataDir := configuration.Settings.GetString("directories.Data")
	indexPath := paths.New(dataDir).Join("library_index.json")
	// Verify the index file exists and we can read its fstat attrs.
	if indexPath.NotExist() {
		return true
	}
	info, err := indexPath.Stat()
	if err != nil {
		return true
	}
	return time.Since(info.ModTime()) > timeout
}
