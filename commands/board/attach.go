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

package board

import (
	"net/url"
	"os"
	"time"

	discovery "github.com/arduino/board-discovery"
	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-modules/boards"
	"github.com/bcmi-labs/arduino-modules/sketches"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	command.AddCommand(attachCommand)
	attachCommand.Flags().StringVar(&attachFlags.boardURI, "board", "", "The URI of the board to connect.")
	attachCommand.Flags().StringVar(&attachFlags.boardFlavour, "flavour", "default", "The Name of the CPU flavour, it is required for some boards (e.g. Arduino Nano).")
	attachCommand.Flags().StringVar(&attachFlags.sketchName, "sketch", "", "The Name of the sketch to attach the board to.")
	attachCommand.Flags().StringVar(&attachFlags.searchTimeout, "timeout", "5s", "The timeout of the search of connected devices, try to high it if your board is not found (e.g. to 10s).")
}

var attachFlags struct {
	boardURI      string // The URI of the board to attach: can be serial:// tty:// http:// https:// tcp:// udp:// referring to the validBoardURIRegexp variable.
	boardFlavour  string // The flavour of the chipset of the cpu of the connected board, if not specified it is set to "default".
	sketchName    string // The name of the sketch to attach to the board.
	fromPath      string // The Path of the file to import and attach to the board.
	searchTimeout string // Expressed in a parsable duration, is the timeout for the list and attach commands.
}

var attachCommand = &cobra.Command{
	Use:   "attach --sketch=[SKETCH-NAME] --board=[BOARD]",
	Short: "Attaches a board to a sketch.",
	Long:  "Attaches a board to a sketch.",
	Example: "" +
		"arduino board attach --board serial:///dev/tty/ACM0 \\\n" +
		"                     --sketch sketchName # Attaches a sketch to a board.",
	Run: runAttachCommand,
}

func runAttachCommand(cmd *cobra.Command, args []string) {
	if attachFlags.sketchName == "" {
		formatter.PrintErrorMessage("No sketch name provided.")
		os.Exit(commands.ErrBadCall)
	}

	if attachFlags.boardURI == "" {
		formatter.PrintErrorMessage("No board URI provided.")
		os.Exit(commands.ErrBadCall)
	}

	duration, err := time.ParseDuration(attachFlags.searchTimeout)
	if err != nil {
		logrus.WithError(err).Warnf("Invalid interval `%s` provided, using default (5s).", attachFlags.searchTimeout)
		duration = time.Second * 5
	}

	monitor := discovery.New(time.Second)
	monitor.Start()

	time.Sleep(duration)

	homeFolder, err := common.GetDefaultArduinoHomeFolder()
	if err != nil {
		formatter.PrintError(err, "Cannot Parse Board Index file.")
		os.Exit(commands.ErrCoreConfig)
	}

	packageFolder, err := common.GetDefaultPkgFolder()
	if err != nil {
		formatter.PrintError(err, "Cannot Parse Board Index file.")
		os.Exit(commands.ErrCoreConfig)
	}

	bs, err := boards.Find(packageFolder)
	if err != nil {
		formatter.PrintError(err, "Cannot Parse Board Index file.")
		os.Exit(commands.ErrCoreConfig)
	}

	ss := sketches.Find(homeFolder)

	sketch, exists := ss[attachFlags.sketchName]
	if !exists {
		formatter.PrintErrorMessage("Cannot find specified sketch in the Sketchbook.")
		os.Exit(commands.ErrGeneric)
	}

	deviceURI, err := url.Parse(attachFlags.boardURI)
	if err != nil {
		formatter.PrintError(err, "The provided Device URL is not in a valid format.")
		os.Exit(commands.ErrBadCall)
	}

	var findBoardFunc func(boards.Boards, *discovery.Monitor, *url.URL) *boards.Board
	var Type string

	if validSerialBoardURIRegexp.Match([]byte(attachFlags.boardURI)) {
		findBoardFunc = findSerialConnectedBoard
		Type = "serial"
	} else if validNetworkBoardURIRegexp.Match([]byte(attachFlags.boardURI)) {
		findBoardFunc = findNetworkConnectedBoard
		Type = "network"
	} else {
		formatter.PrintErrorMessage("Invalid device port type provided. Accepted types are: serial://, tty://, http://, https://, tcp://, udp://.")
		os.Exit(commands.ErrBadCall)
	}

	board := findBoardFunc(bs, monitor, deviceURI)

	sketch.Metadata.CPU = sketches.MetadataCPU{
		Fqbn: board.Fqbn,
		Name: board.Name,
		Type: Type,
	}
	err = sketch.ExportMetadata()
	if err != nil {
		formatter.PrintError(err, "Cannot export sketch metadata.")
	}
	formatter.PrintResult("BOARD ATTACHED.")
}
