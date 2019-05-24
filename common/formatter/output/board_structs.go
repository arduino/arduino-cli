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
