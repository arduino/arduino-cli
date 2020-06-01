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
	"os"
	"path/filepath"

	"github.com/arduino/arduino-cli/i18n/cmd/ast"
	"github.com/spf13/cobra"
)

var generateCatalogCommand = &cobra.Command{
	Use:   "generate [input folder]",
	Short: "generates the en catalog from source files",
	Args:  cobra.MinimumNArgs(1),
	Run:   generateCatalog,
}

func generateCatalog(cmd *cobra.Command, args []string) {

	folder := args[0]
	files := []string{}
	filepath.Walk(folder, func(name string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(name) != ".go" {
			return nil
		}

		files = append(files, name)
		return nil
	})

	catalog := ast.GenerateCatalog(files)
	catalog.Write(os.Stdout)
}
