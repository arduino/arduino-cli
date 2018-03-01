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
	"fmt"
	"net/url"
	"os"
	"regexp"
	"time"

	discovery "github.com/arduino/board-discovery"
	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/bcmi-labs/arduino-modules/boards"
	"github.com/bcmi-labs/arduino-modules/sketches"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var validSerialBoardURIRegexp = regexp.MustCompile("(serial|tty)://.+")
var validNetworkBoardURIRegexp = regexp.MustCompile("(http(s)?|(tc|ud)p)://[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}:[0-9]{1,5}")

func init() {
	command.AddCommand(attachCommand)
	attachCommand.Flags().StringVar(&attachFlags.boardFlavour, "flavour", "default", "The Name of the CPU flavour, it is required for some boards (e.g. Arduino Nano).")
	attachCommand.Flags().StringVar(&attachFlags.searchTimeout, "timeout", "5s", "The timeout of the search of connected devices, try to high it if your board is not found (e.g. to 10s).")
}

var attachFlags struct {
	boardFlavour  string // The flavour of the chipset of the cpu of the connected board, if not specified it is set to "default".
	searchTimeout string // Expressed in a parsable duration, is the timeout for the list and attach commands.
}

var attachCommand = &cobra.Command{
	Use:     "attach sketchName boardURI",
	Short:   "Attaches a sketch to a board.",
	Long:    "Attaches a sketch to a board. Provide sketch name and a board URI to connect.",
	Example: "arduino board attach sketchName serial:///dev/tty/ACM0",
	Args:    cobra.ExactArgs(2),
	Run:     runAttachCommand,
}

func runAttachCommand(cmd *cobra.Command, args []string) {
	sketchName := args[0]
	boardURI := args[1]

	duration, err := time.ParseDuration(attachFlags.searchTimeout)
	if err != nil {
		logrus.WithError(err).Warnf("Invalid interval `%s` provided, using default (5s).", attachFlags.searchTimeout)
		duration = time.Second * 5
	}

	monitor := discovery.New(time.Second)
	monitor.Start()

	time.Sleep(duration)

	// FIXME: Replace with the PackageManager
	homeFolder, err := configs.ArduinoHomeFolder.Get()
	if err != nil {
		formatter.PrintError(err, "Cannot Parse Board Index file.")
		os.Exit(commands.ErrCoreConfig)
	}

	packageFolder, err := configs.PackagesFolder.Get()
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

	sketch, exists := ss[sketchName]
	if !exists {
		formatter.PrintErrorMessage("Cannot find specified sketch in the Sketchbook.")
		os.Exit(commands.ErrGeneric)
	}

	deviceURI, err := url.Parse(boardURI)
	if err != nil {
		formatter.PrintError(err, "The provided Device URL is not in a valid format.")
		os.Exit(commands.ErrBadCall)
	}

	var findBoardFunc func(boards.Boards, *discovery.Monitor, *url.URL) *boards.Board
	var Type string

	if validSerialBoardURIRegexp.Match([]byte(boardURI)) {
		findBoardFunc = findSerialConnectedBoard
		Type = "serial"
	} else if validNetworkBoardURIRegexp.Match([]byte(boardURI)) {
		findBoardFunc = findNetworkConnectedBoard
		Type = "network"
	} else {
		formatter.PrintErrorMessage("Invalid device port type provided. Accepted types are: serial://, tty://, http://, https://, tcp://, udp://.")
		os.Exit(commands.ErrBadCall)
	}

	// TODO: Handle the case when no board is found.
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

// FIXME: Those should probably go in a "BoardManager" pkg or something
// findSerialConnectedBoard find the board which is connected to the specified URI via serial port, using a monitor and a set of Boards
// for the matching.
func findSerialConnectedBoard(bs boards.Boards, monitor *discovery.Monitor, deviceURI *url.URL) *boards.Board {
	found := false
	location := deviceURI.Path
	var serialDevice discovery.SerialDevice
	for _, device := range monitor.Serial() {
		if device.Port == location {
			// Found the device !
			found = true
			serialDevice = *device
		}
	}
	if !found {
		formatter.PrintErrorMessage("No Supported board has been found at the specified board URI.")
		return nil
	}

	board := bs.ByVidPid(serialDevice.VendorID, serialDevice.ProductID)
	if board == nil {
		formatter.PrintErrorMessage("No Supported board has been found, try either install new cores or check your board URI.")
		os.Exit(commands.ErrGeneric)
	}

	formatter.Print("SUPPORTED BOARD FOUND:")
	formatter.Print(board.String())
	return board
}

// findNetworkConnectedBoard find the board which is connected to the specified URI on the network, using a monitor and a set of Boards
// for the matching.
func findNetworkConnectedBoard(bs boards.Boards, monitor *discovery.Monitor, deviceURI *url.URL) *boards.Board {
	found := false

	var networkDevice discovery.NetworkDevice

	for _, device := range monitor.Network() {
		if device.Address == deviceURI.Host &&
			fmt.Sprint(device.Port) == deviceURI.Port() {
			// Found the device !
			found = true
			networkDevice = *device
		}
	}
	if !found {
		formatter.PrintErrorMessage("No Supported board has been found at the specified board URI, try either install new cores or check your board URI.")
		os.Exit(commands.ErrGeneric)
	}

	formatter.Print("SUPPORTED BOARD FOUND:")
	return bs.ByID(networkDevice.Name)
}
