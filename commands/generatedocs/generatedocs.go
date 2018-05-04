/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package generatedocs

import (
	"os"
	"path/filepath"

	"github.com/bcmi-labs/arduino-cli/commands"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// InitCommand prepares the command.
func InitCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:     "generate-docs",
		Short:   "Generates bash completion and command manpages.",
		Long:    "Generates bash completion and command manpages.",
		Example: "arduino generate-docs bash-completions",
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
	command.Hidden = true
	return command
}

var outputDir = ""

func generateBashCompletions(cmd *cobra.Command, args []string) {
	if outputDir == "" {
		outputDir = "docs/bash_completions"
	}
	logrus.WithField("outputDir", outputDir).Info("Generating bash completion")
	err := cmd.Root().GenBashCompletionFile(filepath.Join(outputDir, "arduino"))
	if err != nil {
		logrus.WithError(err).Warn("Error Generating bash autocompletions")
		os.Exit(commands.ErrGeneric)
	}
}

// Generate man pages for all commands and puts them in $PROJECT_DIR/docs/manpages.
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
		os.Exit(commands.ErrGeneric)
	}
}
