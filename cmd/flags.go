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
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
 */

package cmd

// GlobalFlags represents flags available in all the program.
var GlobalFlags struct {
	Verbose int    // More time verbose flag is written, the more the Verbose count increases. Represents verbosity level.
	Format  string // The Output format (e.g. text, json).
	Home    string // The Custom Home directory.
}

// rootCmdFlags represent flags available to the root command.
var rootCmdFlags struct {
	GenerateDocs bool // if true, generates manpages and bash autocompletion.
}

// arduinoLibFlags represents `arduino lib` flags.
var arduinoLibFlags struct {
	updateIndex bool // If true, update library index.
}

// arduinoCoreFlags represents `arduino core` flags.
var arduinoCoreFlags struct {
	updateIndex bool // If true, update package index.
}

// arduinoConfigInitFlags represents `arduino config init` flags.
var arduinoConfigInitFlags struct {
	Default  bool   // If false, ask questions to the user about setting configuration properties, otherwise use default configuration.
	Location string // The custom location of the file to create.
}
