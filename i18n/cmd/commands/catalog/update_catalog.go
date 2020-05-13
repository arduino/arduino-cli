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

package catalog

import (
	"fmt"
	"os"
	"path"

	"github.com/arduino/arduino-cli/i18n/cmd/po"
	"github.com/spf13/cobra"
)

var updateCatalogCommand = &cobra.Command{
	Use:   "update -l pt_BR [catalog folder]",
	Short: "updates the language catalogs base on changes to the en catalog",
	Args:  cobra.ExactArgs(1),
	Run:   updateCatalog,
}

var languages = []string{}

func init() {
	updateCatalogCommand.Flags().StringSliceVarP(&languages, "languages", "l", nil, "languages")
	updateCatalogCommand.MarkFlagRequired("languages")
}

func updateCatalog(cmd *cobra.Command, args []string) {
	folder := args[0]
	enCatalog := po.Parse(path.Join(folder, "en.po"))

	for _, lang := range languages {
		filename := path.Join(folder, fmt.Sprintf("%s.po", lang))
		langCatalog := po.Parse(filename)

		mergedCatalog := po.Merge(enCatalog, langCatalog)

		os.Remove(filename)
		file, err := os.OpenFile(path.Join(folder, fmt.Sprintf("%s.po", lang)), os.O_CREATE|os.O_RDWR, 0644)

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		mergedCatalog.Write(file)
	}
}
