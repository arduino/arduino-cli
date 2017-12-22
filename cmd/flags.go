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

import "github.com/bcmi-labs/arduino-cli/configs"
import "regexp"

// GlobalFlags represents flags available in all the program.
var GlobalFlags struct {
	Debug           bool   // If true, dump debug output to stderr.
	Format          string // The Output format (e.g. text, json).
	configs.Configs        // The Configurations for the CLI.
}

// rootCmdFlags represent flags available to the root command.
var rootCmdFlags struct {
	GenerateDocs bool // if true, generates manpages and bash autocompletion.
}

// arduinoLibFlags represents `arduino lib` flags.
var arduinoLibFlags struct {
	updateIndex bool // if true, updates libraries index.
}

var arduinoLibSearchFlags struct {
	Names bool // if true outputs lib names only.
}

// arduinoCoreFlags represents `arduino core` flags.
var arduinoCoreFlags struct {
	updateIndex bool // If true, update packages index.
}

// arduinoConfigInitFlags represents `arduino config init` flags.
var arduinoConfigInitFlags struct {
	Default  bool   // If false, ask questions to the user about setting configuration properties, otherwise use default configuration.
	Location string // The custom location of the file to create.
}

var validSerialBoardURIRegexp = regexp.MustCompile("(serial|tty)://.+")
var validNetworkBoardURIRegexp = regexp.MustCompile("(http(s)?|(tc|ud)p)://[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}:[0-9]{1,5}")

// arduinoConfigInitFlags represents `arduino board attach` flags.
var arduinoBoardAttachFlags struct {
	BoardURI      string // The URI of the board to attach: can be serial:// tty:// http:// https:// tcp:// udp:// referring to the validBoardURIRegexp variable.
	BoardFlavour  string // The flavour of the chipset of the cpu of the connected board, if not specified it is set to "default".
	SketchName    string // The name of the sketch to attach to the board.
	FromPath      string // The Path of the file to import and attach to the board.
	SearchTimeout string // Expressed in a parsable duration, is the timeout for the list and attach commands
}

// arduinoBoardListFlags represents `arduino board list` flags.
var arduinoBoardListFlags struct {
	SearchTimeout string // Expressed in a parsable duration, is the timeout for the list and attach commands
}

// arduinoSketchSyncFlags represents `arduino sketch sync` flags.
var arduinoSketchSyncFlags struct {
	Priority string // The Prioritary resource when we have conflicts. Can be local, remote, skip-conflict.
}

// arduinoLoginFlags represents `arduino login` flags.
var arduinoLoginFlags struct {
	User     string // The user who asks to login.
	Password string // The password used to authenticate.
}

// arduinoCompileFlags represents `arduino compile` flags.
var arduinoCompileFlags struct {
	FullyQualifiedBoardName string // Fully Qualified Board Name, e.g.: arduino:avr:uno.
	SketchName              string // The name of the sketch to compile.
	DumpPreferences         bool   // Show all build preferences used instead of compiling.
	Preprocess              bool   // Print preprocessed code to stdout.
	BuildPath               string // Folder where to save compiled files.
	Warnings                string // Used to tell gcc which warning level to use.
	Verbose                 bool   // Turns on verbose mode.
	Quiet                   bool   // Supresses almost every output.
	DebugLevel              int    // Used for debugging.
	VidPid                  string // VID/PID specific build properties.
}
