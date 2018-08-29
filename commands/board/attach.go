/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package board

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	discovery "github.com/arduino/board-discovery"
	paths "github.com/arduino/go-paths-helper"
	"github.com/bcmi-labs/arduino-modules/sketches"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initAttachCommand() *cobra.Command {
	attachCommand := &cobra.Command{
		Use:   "attach <port>|<FQBN> [sketchPath]",
		Short: "Attaches a sketch to a board.",
		Long:  "Attaches a sketch to a board.",
		Example: "arduino board attach serial:///dev/tty/ACM0\n" +
			"  " + commands.AppName + " board attach serial:///dev/tty/ACM0 HelloWorld\n" +
			"  " + commands.AppName + " board attach arduino:samd:mkr1000",
		Args: cobra.RangeArgs(1, 2),
		Run:  runAttachCommand,
	}
	attachCommand.Flags().StringVar(&attachFlags.boardFlavour, "flavour", "default", "The Name of the CPU flavour, it is required for some boards (e.g. Arduino Nano).")
	attachCommand.Flags().StringVar(&attachFlags.searchTimeout, "timeout", "5s", "The timeout of the search of connected devices, try to high it if your board is not found (e.g. to 10s).")
	return attachCommand
}

var attachFlags struct {
	boardFlavour  string // The flavor of the chipset of the cpu of the connected board, if not specified it is set to "default".
	searchTimeout string // Expressed in a parsable duration, is the timeout for the list and attach commands.
}

func runAttachCommand(cmd *cobra.Command, args []string) {
	boardURI := args[0]
	var sketchPath *paths.Path
	if len(args) > 1 {
		sketchPath = paths.New(args[1])
	}
	sketch, err := commands.InitSketch(sketchPath)
	if err != nil {
		formatter.PrintError(err, "Error opening sketch.")
		os.Exit(commands.ErrGeneric)
	}

	logrus.WithField("fqbn", boardURI).Print("Parsing FQBN")
	fqbn, err := cores.ParseFQBN(boardURI)
	if err != nil && !strings.HasPrefix(boardURI, "serial") {
		boardURI = "serial://" + boardURI
	}

	pm := commands.InitPackageManager()

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
			formatter.PrintErrorMessage("No supported board has been found at " + deviceURI.String() + ", try either install new cores or check your board URI.")
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
		return nil
	}

	boards := pm.FindBoardsWithVidPid(serialDevice.VendorID, serialDevice.ProductID)
	if len(boards) == 0 {
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
		return nil
	}

	boards := pm.FindBoardsWithID(networkDevice.Name)
	if len(boards) == 0 {
		return nil
	}

	return boards[0]
}
