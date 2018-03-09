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

package sketch

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/spf13/cobra"
)

func init() {
	command.AddCommand(newCommand)
}

var newCommand = &cobra.Command{
	Use:     "new",
	Short:   "Create a new Sketch",
	Long:    "Create a new Sketch",
	Example: "arduino sketch new MultiBlinker",
	Args:    cobra.ExactArgs(1),
	Run:     runNewCommand,
}

var emptySketch = []byte(`
void setup() {
}

void loop() {
}
`)

func runNewCommand(cmd *cobra.Command, args []string) {
	sketchbook, err := configs.ArduinoHomeFolder.Get()
	if err != nil {
		formatter.PrintError(err, "Cannot get sketchbook folder.")
		os.Exit(commands.ErrCoreConfig)
	}

	sketchDir := filepath.Join(sketchbook, args[0])
	if err := os.Mkdir(sketchDir, 0755); err != nil {
		formatter.PrintError(err, "Could not create sketch folder.")
		os.Exit(commands.ErrGeneric)
	}

	sketchFile := filepath.Join(sketchDir, args[0]+".ino")
	if err := ioutil.WriteFile(sketchFile, emptySketch, 0644); err != nil {
		formatter.PrintError(err, "Error creating sketch.")
		os.Exit(commands.ErrGeneric)
	}

	formatter.Print("Sketch created in: " + sketchDir)
}
