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

	"github.com/gosuri/uitable"
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

// AttachedBoardList is a list of attached boards.
type AttachedBoardList struct {
	SerialBoards  []SerialBoardListItem  `json:"serialBoards,required"`
	NetworkBoards []NetworkBoardListItem `json:"networkBoards,required"`
}

func (bl *AttachedBoardList) String() string {
	table := uitable.New()
	table.MaxColWidth = 100
	table.Wrap = true // wrap columns

	table.AddRow("FQBN", "Port", "ID", "Board Name")
	for _, item := range bl.SerialBoards {
		table.AddRow(item.Fqbn, item.Port, item.UsbID[:9], item.Name)
	}
	for _, item := range bl.NetworkBoards {
		table.AddRow(item.Fqbn, "network://"+item.Location, "", item.Name)
	}
	return fmt.Sprintln(table)
}

// BoardListItem is a supported board
type BoardListItem struct {
	Name string `json:"name,required"`
	Fqbn string `json:"fqbn,required"`
}

// BoardList is a list of supported boards
type BoardList struct {
	Boards []*BoardListItem `json:"boards,required"`
}

func (bl *BoardList) String() string {
	table := uitable.New()
	table.MaxColWidth = 100
	table.Wrap = true // wrap columns

	table.AddRow("Board Name", "FQBN")
	for _, item := range bl.Boards {
		table.AddRow(item.Name, item.Fqbn)
	}
	return fmt.Sprintln(table)
}

func (bl *BoardList) Len() int {
	return len(bl.Boards)
}

func (bl *BoardList) Less(i, j int) bool {
	return bl.Boards[i].Name < bl.Boards[j].Name
}

func (bl *BoardList) Swap(i, j int) {
	bl.Boards[i], bl.Boards[j] = bl.Boards[j], bl.Boards[i]
}
