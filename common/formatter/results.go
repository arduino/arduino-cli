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
	"encoding/json"
	"fmt"
)

type resultMessage struct {
	message interface{}
}

// MarshalJSON allows to marshal this object as a JSON object.
func (res resultMessage) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"result": res.message,
	})
}

func (res resultMessage) String() string {
	return fmt.Sprint(res.message)
}

// PrintResult prints a value as a result from an operation.
func PrintResult(res interface{}) {
	Print(resultMessage{
		message: res,
	})
}
