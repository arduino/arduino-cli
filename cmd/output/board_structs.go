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

package output

import (
	"fmt"

	discovery "github.com/arduino/board-discovery"
	"github.com/bcmi-labs/arduino-modules/boards"
)

// SerialBoardListItem represents a board connected using serial port.
type SerialBoardListItem struct {
	Name  string `json:"name,required"`
	Fqbn  string `json:"fqbn,required"`
	Port  string `json:"port,required"`
	UsbID string `json:"usbID,reqiured"`
}

// NetworkBoardListItem represents a board connected via network.
type NetworkBoardListItem struct {
	Name     string `json:"name,required"`
	Fqbn     string `json:"fqbn,required"`
	Location string `json:"location,required"`
}

// BoardList is a list of attached boards.
type BoardList struct {
	SerialBoards  []SerialBoardListItem  `json:"serialBoards,required"`
	NetworkBoards []NetworkBoardListItem `json:"networkBoards,required"`
}

func (bl *BoardList) String() string {
	ret := fmt.Sprintln("DEVICES:") +
		fmt.Sprintln("SERIAL:")
	if len(bl.SerialBoards) == 0 {
		ret += fmt.Sprintln("   <none>")
	} else {
		for _, item := range bl.SerialBoards {
			ret += fmt.Sprintln(" - BOARD NAME:", item.Name) +
				fmt.Sprintln("   FQBN:", item.Fqbn) +
				fmt.Sprintln("   PORT:", item.Port) +
				fmt.Sprintln("   USB ID:", item.UsbID)
		}
	}
	ret += fmt.Sprintln("NETWORK:")
	if len(bl.NetworkBoards) == 0 {
		ret += fmt.Sprintln("   <none>")
	} else {
		for _, item := range bl.NetworkBoards {
			ret += fmt.Sprintln(" - BOARD NAME:", item.Name) +
				fmt.Sprintln("   FQBN:", item.Fqbn) +
				fmt.Sprintln("   LOCATION:", item.Location)
		}
	}
	return ret
}

// NewBoardList returns a new board list by adding discovered boards from the board list and a monitor.
func NewBoardList(boards boards.Boards, monitor *discovery.Monitor) *BoardList {
	if monitor == nil || boards == nil {
		return nil
	}

	serialDevices := monitor.Serial()
	networkDevices := monitor.Network()
	ret := &BoardList{
		SerialBoards:  make([]SerialBoardListItem, 0, len(serialDevices)),
		NetworkBoards: make([]NetworkBoardListItem, 0, len(networkDevices)),
	}

	for _, item := range serialDevices {
		board := boards.ByVidPid(item.VendorID, item.ProductID)
		if board == nil { // skip it if not recognized
			continue
		}

		ret.SerialBoards = append(ret.SerialBoards, SerialBoardListItem{
			Name:  board.Name,
			Fqbn:  board.Fqbn,
			Port:  item.Port,
			UsbID: fmt.Sprintf("%s:%s - %s", item.ProductID[2:len(item.ProductID)-1], item.VendorID[2:len(item.VendorID)-1], item.SerialNumber),
		})
	}

	for _, item := range networkDevices {
		board := boards.ByID(item.Name)
		if board == nil { // skip it if not recognized
			continue
		}

		ret.NetworkBoards = append(ret.NetworkBoards, NetworkBoardListItem{
			Name:     board.Name,
			Fqbn:     board.Fqbn,
			Location: fmt.Sprintf("%s:%d", item.Address, item.Port),
		})
	}
	return ret
}
