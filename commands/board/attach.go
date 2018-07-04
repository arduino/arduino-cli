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
	"strings"
	"time"

	discovery "github.com/arduino/board-discovery"
	paths "github.com/arduino/go-paths-helper"
	"github.com/bcmi-labs/arduino-cli/arduino/cores"
	"github.com/bcmi-labs/arduino-cli/arduino/cores/packagemanager"
	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-modules/sketches"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initAttachCommand() *cobra.Command {
	attachCommand := &cobra.Command{
		Use:     "attach <port>|<FQBN> [sketchPath]",
		Short:   "Attaches a sketch to a board.",
		Long:    "Attaches a sketch to a board.",
		Example: "arduino board attach serial:///dev/tty/ACM0",
		Args:    cobra.RangeArgs(1, 2),
		Run:     runAttachCommand,
	}
	attachCommand.Flags().StringVar(&attachFlags.boardFlavour, "flavour", "default", "The Name of the CPU flavour, it is required for some boards (e.g. Arduino Nano).")
	attachCommand.Flags().StringVar(&attachFlags.searchTimeout, "timeout", "5s", "The timeout of the search of connected devices, try to high it if your board is not found (e.g. to 10s).")
	return attachCommand
}

var attachFlags struct {
	boardFlavour  string // The flavour of the chipset of the cpu of the connected board, if not specified it is set to "default".
	searchTimeout string // Expressed in a parsable duration, is the timeout for the list and attach commands.
}

func runAttachCommand(cmd *cobra.Command, args []string) {
	boardURI := args[0]
	var sketchPath *paths.Path
	if len(args) > 0 {
		sketchPath = paths.New(args[0])
	}
	sketch, err := commands.InitSketch(sketchPath)
	if err != nil {
		formatter.PrintError(err, "Error opening sketch.")
		os.Exit(commands.ErrGeneric)
	}

	fqbn, err := cores.ParseFQBN(boardURI)
	if err != nil && !strings.HasPrefix(boardURI, "serial") {
		boardURI = "serial://" + boardURI
	}

	pm := commands.InitPackageManager()
	if err = pm.LoadHardware(); err != nil {
		formatter.PrintError(err, "Error loading hardware packages.")
		os.Exit(commands.ErrCoreConfig)
	}

	if fqbn != nil {
		sketch.Metadata.CPU = sketches.MetadataCPU{
			Fqbn: fqbn.String(),
		}
	} else {
		deviceURI, err := url.Parse(boardURI)
		if err != nil {
			formatter.PrintError(err, "The provided Device URL is not in a valid format.")
			os.Exit(commands.ErrBadCall)
		}

		var findBoardFunc func(*packagemanager.PackageManager, *discovery.Monitor, *url.URL) *cores.Board
		var Type string
		switch deviceURI.Scheme {
		case "serial", "tty":
			findBoardFunc = findSerialConnectedBoard
			Type = "serial"
		case "http", "https", "tcp", "udp":
			findBoardFunc = findNetworkConnectedBoard
			Type = "network"
		default:
			formatter.PrintErrorMessage("Invalid device port type provided. Accepted types are: serial://, tty://, http://, https://, tcp://, udp://.")
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

		// TODO: Handle the case when no board is found.
		board := findBoardFunc(pm, monitor, deviceURI)
		if board == nil {
			os.Exit(commands.ErrGeneric)
		}
		formatter.Print("Board found: " + board.Name())

		sketch.Metadata.CPU = sketches.MetadataCPU{
			Fqbn: board.FQBN(),
			Name: board.Name(),
			Type: Type,
		}
	}

	err = sketch.ExportMetadata()
	if err != nil {
		formatter.PrintError(err, "Cannot export sketch metadata.")
	}
	formatter.PrintResult("Selected fqbn: " + sketch.Metadata.CPU.Fqbn)
}

// FIXME: Those should probably go in a "BoardManager" pkg or something
// findSerialConnectedBoard find the board which is connected to the specified URI via serial port, using a monitor and a set of Boards
// for the matching.
func findSerialConnectedBoard(pm *packagemanager.PackageManager, monitor *discovery.Monitor, deviceURI *url.URL) *cores.Board {
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
		formatter.PrintErrorMessage("Sorry, no Supported board has been found at the specified board URI :-(")
		return nil
	}

	boards := pm.FindBoardsWithVidPid(serialDevice.VendorID, serialDevice.ProductID)
	if len(boards) == 0 {
		formatter.PrintErrorMessage("No Supported board has been found, try either install new cores or check your board URI.")
		os.Exit(commands.ErrGeneric)
	}

	return boards[0]
}

// findNetworkConnectedBoard find the board which is connected to the specified URI on the network, using a monitor and a set of Boards
// for the matching.
func findNetworkConnectedBoard(pm *packagemanager.PackageManager, monitor *discovery.Monitor, deviceURI *url.URL) *cores.Board {
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

	boards := pm.FindBoardsWithID(networkDevice.Name)
	if len(boards) == 0 {
		formatter.PrintErrorMessage("No Supported board has been found, try either install new cores or check your board URI.")
		os.Exit(commands.ErrGeneric)
	}

	return boards[0]
}
