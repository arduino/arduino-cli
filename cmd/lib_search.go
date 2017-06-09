// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"errors"
	"fmt"

	"strings"

	"github.com/arduino/arduino-cli/libraries"
	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Searchs for a library data",
	Long:  `Search for one or more libraries data.`,
	RunE:  executeSearch,
}

func init() {
	LibRoot.AddCommand(searchCmd)
}

func executeSearch(cmd *cobra.Command, args []string) error {
	query := ""
	if len(args) > 1 {
		return errors.New("Wrong Number of Arguments")
	}
	if len(args) == 1 {
		query = args[0]
	}

	index, err := libraries.LoadLibrariesIndex()
	if err != nil {
		fmt.Println("Index file is corrupted. Please replace it by updating : arduino lib list update")
		return nil
	}

	libraries, err := libraries.CreateStatusContextFromIndex(index, nil, nil)
	if err != nil {
		fmt.Printf("Could not synchronize library status: %s", err)
		return nil
	}

	found := false

	//Pretty print libraries from index.
	for _, name := range libraries.Names() {
		if strings.Contains(name, query) {
			found = true
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

	if !found {
		fmt.Printf("No library found matching \"%s\" search query.\n", query)
	}

	return nil
}
