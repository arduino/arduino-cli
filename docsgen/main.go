// Source: https://github.com/arduino/tooling-project-assets/blob/main/workflow-templates/assets/cobra/docsgen/main.go
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

// Package main generates Markdown documentation for the project's CLI.
package main

import (
	"os"

	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/arduino-cli/internal/cli"
	"github.com/spf13/cobra/doc"
)

func main() {
	if len(os.Args) < 2 {
		print("error: Please provide the output folder argument")
		os.Exit(1)
	}

	os.MkdirAll(os.Args[1], 0755) // Create the output folder if it doesn't already exist

	configuration.Settings = configuration.Init(configuration.FindConfigFileInArgsOrWorkingDirectory(os.Args))
	cli := cli.NewCommand()
	cli.DisableAutoGenTag = true // Disable addition of auto-generated date stamp
	err := doc.GenMarkdownTree(cli, os.Args[1])
	if err != nil {
		panic(err)
	}
}
