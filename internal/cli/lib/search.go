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
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/feedback/result"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initSearchCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	var namesOnly bool
	var omitReleasesDetails bool
	searchCommand := &cobra.Command{
		Use:   fmt.Sprintf("search [%s ...]", i18n.Tr("SEARCH_TERM")),
		Short: i18n.Tr("Searches for one or more libraries matching a query."),
		Long: i18n.Tr(`Search for libraries matching zero or more search terms.

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
		Example: "  " + os.Args[0] + " lib search audio                               # " + i18n.Tr("basic search for \"audio\"") + "\n" +
			"  " + os.Args[0] + " lib search name:buzzer                         # " + i18n.Tr("libraries with \"buzzer\" in the Name field") + "\n" +
			"  " + os.Args[0] + " lib search name=pcf8523                        # " + i18n.Tr("libraries with a Name exactly matching \"pcf8523\"") + "\n" +
			"  " + os.Args[0] + " lib search \"author:\\\"Daniel Garcia\\\"\"          # " + i18n.Tr("libraries authored by Daniel Garcia") + "\n" +
			"  " + os.Args[0] + " lib search author=Adafruit name:gfx            # " + i18n.Tr("libraries authored only by Adafruit with \"gfx\" in their Name") + "\n" +
			"  " + os.Args[0] + " lib search esp32 display maintainer=espressif  # " + i18n.Tr("basic search for \"esp32\" and \"display\" limited to official Maintainer") + "\n" +
			"  " + os.Args[0] + " lib search dependencies:IRremote               # " + i18n.Tr("libraries that depend on at least \"IRremote\"") + "\n" +
			"  " + os.Args[0] + " lib search dependencies=IRremote               # " + i18n.Tr("libraries that depend only on \"IRremote\"") + "\n",
		Args: cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runSearchCommand(cmd.Context(), srv, args, namesOnly, omitReleasesDetails)
		},
	}
	searchCommand.Flags().BoolVar(&namesOnly, "names", false, i18n.Tr("Show library names only."))
	searchCommand.Flags().BoolVar(&omitReleasesDetails, "omit-releases-details", false, i18n.Tr("Omit library details far all versions except the latest (produce a more compact JSON output)."))
	return searchCommand
}

// indexUpdateInterval specifies the time threshold over which indexes are updated
const indexUpdateInterval = 60 * time.Minute

func runSearchCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer, args []string, namesOnly bool, omitReleasesDetails bool) {
	inst := instance.CreateAndInit(ctx, srv)

	logrus.Info("Executing `arduino-cli lib search`")

	stream, res := commands.UpdateLibrariesIndexStreamResponseToCallbackFunction(ctx, feedback.ProgressBar())
	req := &rpc.UpdateLibrariesIndexRequest{Instance: inst, UpdateIfOlderThanSecs: int64(indexUpdateInterval.Seconds())}
	if err := srv.UpdateLibrariesIndex(req, stream); err != nil {
		feedback.Fatal(i18n.Tr("Error updating library index: %v", err), feedback.ErrGeneric)
	}
	if res().GetLibrariesIndex().GetStatus() == rpc.IndexUpdateReport_STATUS_UPDATED {
		instance.Init(ctx, srv, inst)
	}

	// Perform library search
	searchResp, err := srv.LibrarySearch(ctx, &rpc.LibrarySearchRequest{
		Instance:            inst,
		SearchArgs:          strings.Join(args, " "),
		OmitReleasesDetails: omitReleasesDetails,
	})
	if err != nil {
		feedback.Fatal(i18n.Tr("Error searching for Libraries: %v", err), feedback.ErrGeneric)
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
			names = append(names, LibName{lib.Name})
		}

		return NamesOnly{names}
	}

	return res.results
}

func (res librarySearchResult) String() string {
	results := res.results.Libraries
	if len(results) == 0 {
		return i18n.Tr("No libraries matching your search.")
	}

	var out strings.Builder

	if res.results.Status == result.LibrarySearchStatusFailed {
		out.WriteString(i18n.Tr("No libraries matching your search.\nDid you mean...\n"))
	}

	for _, lib := range results {
		if res.results.Status == result.LibrarySearchStatusSuccess {
			out.WriteString(i18n.Tr(`Name: "%s"`, lib.Name) + "\n")
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
			if dep.VersionConstraint == "" {
				deps = append(deps, dep.Name)
			} else {
				deps = append(deps, dep.Name+" ("+dep.VersionConstraint+")")
			}
		}

		out.WriteString(fmt.Sprintf("  "+i18n.Tr("Author: %s")+"\n", latest.Author))
		out.WriteString(fmt.Sprintf("  "+i18n.Tr("Maintainer: %s")+"\n", latest.Maintainer))
		out.WriteString(fmt.Sprintf("  "+i18n.Tr("Sentence: %s")+"\n", latest.Sentence))
		out.WriteString(fmt.Sprintf("  "+i18n.Tr("Paragraph: %s")+"\n", latest.Paragraph))
		out.WriteString(fmt.Sprintf("  "+i18n.Tr("Website: %s")+"\n", latest.Website))
		if latest.License != "" {
			out.WriteString(fmt.Sprintf("  "+i18n.Tr("License: %s")+"\n", latest.License))
		}
		out.WriteString(fmt.Sprintf("  "+i18n.Tr("Category: %s")+"\n", latest.Category))
		out.WriteString(fmt.Sprintf("  "+i18n.Tr("Architecture: %s")+"\n", strings.Join(latest.Architectures, ", ")))
		out.WriteString(fmt.Sprintf("  "+i18n.Tr("Types: %s")+"\n", strings.Join(latest.Types, ", ")))
		out.WriteString(fmt.Sprintf("  "+i18n.Tr("Versions: %s")+"\n", strings.ReplaceAll(fmt.Sprint(lib.AvailableVersions), " ", ", ")))
		if len(latest.ProvidesIncludes) > 0 {
			out.WriteString(fmt.Sprintf("  "+i18n.Tr("Provides includes: %s")+"\n", strings.Join(latest.ProvidesIncludes, ", ")))
		}
		if len(latest.Dependencies) > 0 {
			out.WriteString(fmt.Sprintf("  "+i18n.Tr("Dependencies: %s")+"\n", strings.Join(deps, ", ")))
		}
	}

	return out.String()
}
