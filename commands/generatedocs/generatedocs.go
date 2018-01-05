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
	"fmt"

	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// Init prepares the command.
func Init(rootCommand *cobra.Command) {
	rootCommand.AddCommand(command)
}

var command = &cobra.Command{
	Use:     "generate-docs",
	Short:   "Generates documentation.",
	Long:    "Generates bash autocompletion, command manpages and puts it into the docs folder.",
	Example: "arduino generate-docs",
	Args:    cobra.NoArgs,
	Run:     run,
}

func run(cmd *cobra.Command, args []string) {
	logrus.Info("Generating docs")
	errorText := ""
	err := cmd.Parent().GenBashCompletionFile("docs/bash_completions/arduino")
	if err != nil {
		logrus.WithError(err).Warn("Error Generating bash autocompletions")
		errorText += fmt.Sprintln(err)
	}
	err = generateManPages(cmd.Parent())
	if err != nil {
		logrus.WithError(err).Warn("Error Generating manpages")
		errorText += fmt.Sprintln(err)
	}
	if errorText != "" {
		formatter.PrintErrorMessage(errorText)
	}
}

// Generate man pages for all commands and puts them in $PROJECT_DIR/docs/manpages.
func generateManPages(rootCommand *cobra.Command) error {
	header := &doc.GenManHeader{
		Title:   "ARDUINO COMMAND LINE MANUAL",
		Section: "1",
	}
	return doc.GenManTree(rootCommand, header, "./docs/manpages")
}
