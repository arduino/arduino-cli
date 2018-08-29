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
	"time"

	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/common/formatter/output"
	"github.com/arduino/board-discovery"
	"github.com/codeclysm/cc"
	"github.com/spf13/cobra"
)

func initListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:     "list",
		Short:   "List connected boards.",
		Long:    "Detects and displays a list of connected boards to the current computer.",
		Example: "  " + commands.AppName + " board list --timeout 10s",
		Args:    cobra.NoArgs,
		Run:     runListCommand,
	}
	usage := "The timeout of the search of connected devices, try to high it if your board is not found (e.g. to 10s)."
	listCommand.Flags().StringVar(&listFlags.timeout, "timeout", "5s", usage)
	return listCommand
}

var listFlags struct {
	timeout string // Expressed in a parsable duration, is the timeout for the list and attach commands.
}

// runListCommand detects and lists the connected arduino boards
// (either via serial or network ports).
func runListCommand(cmd *cobra.Command, args []string) {
	pm := commands.InitPackageManager()

	monitor := discovery.New(time.Millisecond)
	monitor.Start()
	duration, err := time.ParseDuration(listFlags.timeout)
	if err != nil {
		duration = time.Second * 5
	}
	if formatter.IsCurrentFormat("text") {
		stoppable := cc.Run(func(stop chan struct{}) {
			for {
				select {
				case <-stop:
					fmt.Print("\r              \r")
					return
				default:
					fmt.Print("\rDiscovering.  ")
					time.Sleep(time.Millisecond * 500)
					fmt.Print("\rDiscovering.. ")
					time.Sleep(time.Millisecond * 500)
					fmt.Print("\rDiscovering...")
					time.Sleep(time.Millisecond * 500)
				}
			}
		})

		fmt.Print("\r")

		time.Sleep(duration)
		stoppable.Stop()
		<-stoppable.Stopped
	} else {
		time.Sleep(duration)
	}

	formatter.Print(NewBoardList(pm, monitor))

	//monitor.Stop() //If called will slow like 1sec the program to close after print, with the same result (tested).
	// it closes ungracefully, but at the end of the command we can't have races.
}

// NewBoardList returns a new board list by adding discovered boards from the board list and a monitor.
func NewBoardList(pm *packagemanager.PackageManager, monitor *discovery.Monitor) *output.AttachedBoardList {
	if monitor == nil {
		return nil
	}

	serialDevices := monitor.Serial()
	networkDevices := monitor.Network()
	ret := &output.AttachedBoardList{
		SerialBoards:  make([]output.SerialBoardListItem, 0, len(serialDevices)),
		NetworkBoards: make([]output.NetworkBoardListItem, 0, len(networkDevices)),
	}

	for _, item := range serialDevices {
		boards := pm.FindBoardsWithVidPid(item.VendorID, item.ProductID)
		if len(boards) == 0 {
			ret.SerialBoards = append(ret.SerialBoards, output.SerialBoardListItem{
				Name:  "unknown",
				Port:  item.Port,
				UsbID: fmt.Sprintf("%s:%s - %s", item.VendorID[2:], item.ProductID[2:], item.SerialNumber),
			})
			continue
		}

		board := boards[0]
		ret.SerialBoards = append(ret.SerialBoards, output.SerialBoardListItem{
			Name:  board.Name(),
			Fqbn:  board.FQBN(),
			Port:  item.Port,
			UsbID: fmt.Sprintf("%s:%s - %s", item.VendorID[2:], item.ProductID[2:], item.SerialNumber),
		})
	}

	for _, item := range networkDevices {
		boards := pm.FindBoardsWithID(item.Name)
		if len(boards) == 0 {
			// skip it if not recognized
			continue
		}

		board := boards[0]
		ret.NetworkBoards = append(ret.NetworkBoards, output.NetworkBoardListItem{
			Name:     board.Name(),
			Fqbn:     board.FQBN(),
			Location: fmt.Sprintf("%s:%d", item.Address, item.Port),
		})
	}
	return ret
}
