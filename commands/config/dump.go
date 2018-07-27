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
 * Copyright 2017-2018 ARDUINO AG (http://www.arduino.cc/)
 */

package config

import (
	"fmt"
	"os"

	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initDumpCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "dump",
		Short:   "Prints the current configuration",
		Long:    "Prints the current configuration.",
		Example: "arduino config dump",
		Args:    cobra.NoArgs,
		Run:     runDumpCommand,
	}
}

var dumpFlags struct {
	_default bool   // If false, ask questions to the user about setting configuration properties, otherwise use default configuration.
	location string // The custom location of the file to create.
}

func runDumpCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino config dump`")

	data, err := commands.Config.SerializeToYAML()
	if err != nil {
		formatter.PrintError(err, "Error creating configuration")
		os.Exit(commands.ErrGeneric)
	}

	fmt.Println(string(data))
}
