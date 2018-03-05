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
