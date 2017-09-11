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

package cmd

import (
	"fmt"

	"github.com/bcmi-labs/arduino-cli/cmd/formatter"
	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/spf13/cobra"
)

var arduinoConfigCmd = &cobra.Command{
	Use:     `config`,
	Short:   `Arduino Configuration Commands`,
	Long:    `Arduino Configuration Commands`,
	Example: `arduino config init # Initializes a new config file into the default location`,
}

var arduinoConfigInitCmd = &cobra.Command{
	Use:   `init`,
	Short: `Initializes a new config file into the default location`,
	Long:  `Initializes a new config file into the default location ($EXE_DIR/cli-config.yml)`,
	Example: `arduino config init           # Creates a config file by asking questions to the user into the default location
arduino config init --default # Creates a config file with default configuration into default location`,
	Run: executeConfigInitCommand,
}

func executeConfigInitCommand(cmd *cobra.Command, args []string) {
	var conf configs.Configs
	if !arduinoConfigInitFlags.Default && formatter.IsCurrentFormat("text") {
		conf = ConfigsFromQuestions()
	} else {
		conf = configs.Default()
	}
	err := conf.Serialize(arduinoConfigInitFlags.Location)
	if err != nil {
		formatter.PrintErrorMessage(fmt.Sprint("Config file creation error: ", err))
	} else {
		formatter.PrintResult("Config file PATH: " + arduinoConfigInitFlags.Location)
	}
}

// ConfigsFromQuestions asks some questions to the user to properly initialize configs.
// It does not have much sense to use it in JSON formatting, though.
func ConfigsFromQuestions() configs.Configs {
	ret := configs.Default()
	//Set of questions here
	return ret
}
