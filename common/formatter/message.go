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

package formatter

import (
	"fmt"
	"strings"
)

// Message represents a formattable message.
type Message struct {
	Header string      `json:"header,omitempty"` // What is written before parsing Data.
	Data   interface{} `json:"data,omitempty"`   // The Data of the message, this should be the most important data to convert.
	Footer string      `json:"footer,omitempty"` // What is written after parsing Data.
}

// String returns a string representation of the object.
func (m *Message) String() string {
	data := fmt.Sprintf("%s", m.Data)
	message := m.Header
	if message != "" {
		message += "\n"
	}
	message += data
	if data != "" {
		message += "\n"
	}
	return strings.TrimSpace(message + m.Footer)
}
