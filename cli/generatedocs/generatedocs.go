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

package generatedocs

import (
	"os"
	"path/filepath"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var outputDir = ""

// NewCommand created a new `generatedocs` command
func NewCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "generate-docs",
		Short:   "Generates bash completion and command manpages.",
		Long:    "Generates bash completion and command manpages.",
		Example: "  " + os.Args[0] + " generate-docs bash-completions",
		Hidden:  true,
	}

	command.PersistentFlags().StringVarP(&outputDir, "output-dir", "o", "",
		"Directory where to save generated files. Default is './docs', the directory must exist.")
	command.AddCommand(&cobra.Command{
		Use:  "manpage",
		Args: cobra.NoArgs,
		Run:  generateManPages,
	})
	command.AddCommand(&cobra.Command{
		Use:  "bash-completions",
		Args: cobra.NoArgs,
		Run:  generateBashCompletions,
	})

	return command
}

func generateBashCompletions(cmd *cobra.Command, args []string) {
	if outputDir == "" {
		outputDir = "docs/bash_completions"
	}
	logrus.WithField("outputDir", outputDir).Info("Generating bash completion")
	err := cmd.Root().GenBashCompletionFile(filepath.Join(outputDir, "arduino"))
	if err != nil {
		logrus.WithError(err).Warn("Error Generating bash autocompletions")
		os.Exit(errorcodes.ErrGeneric)
	}
}

// Generates man pages for all commands and puts them in $PROJECT_DIR/docs/manpages.
func generateManPages(cmd *cobra.Command, args []string) {
	if outputDir == "" {
		outputDir = "docs/manpages"
	}
	logrus.WithField("outputDir", outputDir).Info("Generating manpages")
	header := &doc.GenManHeader{
		Title:   "ARDUINO COMMAND LINE MANUAL",
		Section: "1",
	}
	err := doc.GenManTree(cmd.Root(), header, outputDir)
	if err != nil {
		logrus.WithError(err).Warn("Error Generating manpages")
		os.Exit(errorcodes.ErrGeneric)
	}
}
