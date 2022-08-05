// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package board

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	discovery "github.com/arduino/board-discovery"
	"github.com/arduino/go-paths-helper"
)

var tr = i18n.Tr

// Attach FIXMEDOC
func Attach(ctx context.Context, req *rpc.BoardAttachRequest, taskCB rpc.TaskProgressCB) (*rpc.BoardAttachResponse, error) {
	pme, release := commands.GetPackageManagerExplorer(req)
	if pme == nil {
		return nil, &arduino.InvalidInstanceError{}
	}
	defer release()
	var sketchPath *paths.Path
	if req.GetSketchPath() != "" {
		sketchPath = paths.New(req.GetSketchPath())
	}
	sk, err := sketch.New(sketchPath)
	if err != nil {
		return nil, &arduino.CantOpenSketchError{Cause: err}
	}

	boardURI := req.GetBoardUri()
	fqbn, err := cores.ParseFQBN(boardURI)
	if err != nil && !strings.HasPrefix(boardURI, "serial") {
		boardURI = "serial://" + boardURI
	}

	if fqbn != nil {
		sk.Metadata.CPU = sketch.BoardMetadata{
			Fqbn: fqbn.String(),
		}
	} else {
		deviceURI, err := url.Parse(boardURI)
		if err != nil {
			return nil, &arduino.InvalidArgumentError{Message: tr("Invalid Device URL format"), Cause: err}
		}

		var findBoardFunc func(*packagemanager.Explorer, *discovery.Monitor, *url.URL) *cores.Board
		switch deviceURI.Scheme {
		case "serial", "tty":
			findBoardFunc = findSerialConnectedBoard
		case "http", "https", "tcp", "udp":
			findBoardFunc = findNetworkConnectedBoard
		default:
			return nil, &arduino.InvalidArgumentError{Message: tr("Invalid device port type provided")}
		}

		duration, err := time.ParseDuration(req.GetSearchTimeout())
		if err != nil {
			duration = time.Second * 5
		}

		monitor := discovery.New(time.Second)
		monitor.Start()

		time.Sleep(duration)

		// TODO: Handle the case when no board is found.
		board := findBoardFunc(pme, monitor, deviceURI)
		if board == nil {
			return nil, &arduino.InvalidArgumentError{Message: tr("No supported board found at %s", deviceURI)}
		}
		taskCB(&rpc.TaskProgress{Name: tr("Board found: %s", board.Name())})

		// TODO: should be stoped the monitor: when running as a pure CLI  is released
		// by the OS, when run as daemon the resource's state is unknown and could be leaked.
		sk.Metadata.CPU = sketch.BoardMetadata{
			Fqbn: board.FQBN(),
			Name: board.Name(),
			Port: deviceURI.String(),
		}
	}

	err = sk.ExportMetadata()
	if err != nil {
		return nil, &arduino.PermissionDeniedError{Message: tr("Cannot export sketch metadata"), Cause: err}
	}
	taskCB(&rpc.TaskProgress{Name: tr("Selected fqbn: %s", sk.Metadata.CPU.Fqbn), Completed: true})
	return &rpc.BoardAttachResponse{}, nil
}

// FIXME: Those should probably go in a "BoardManager" pkg or something
// findSerialConnectedBoard find the board which is connected to the specified URI via serial port, using a monitor and a set of Boards
// for the matching.
func findSerialConnectedBoard(pme *packagemanager.Explorer, monitor *discovery.Monitor, deviceURI *url.URL) *cores.Board {
	found := false
	// to support both cases:
	// serial:///dev/ttyACM2 parsing gives: deviceURI.Host = ""      and deviceURI.Path = /dev/ttyACM2
	// serial://COM3 parsing gives:         deviceURI.Host = "COM3"  and deviceURI.Path = ""
	location := deviceURI.Host + deviceURI.Path
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

	boards := pme.FindBoardsWithVidPid(serialDevice.VendorID, serialDevice.ProductID)
	if len(boards) == 0 {
		return nil
	}

	return boards[0]
}

// findNetworkConnectedBoard find the board which is connected to the specified URI on the network, using a monitor and a set of Boards
// for the matching.
func findNetworkConnectedBoard(pme *packagemanager.Explorer, monitor *discovery.Monitor, deviceURI *url.URL) *cores.Board {
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

	boards := pme.FindBoardsWithID(networkDevice.Name)
	if len(boards) == 0 {
		return nil
	}

	return boards[0]
}
