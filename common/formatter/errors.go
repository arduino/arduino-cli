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
	"strings"
)

// ErrorMessage represents an Error with an attached message.
type ErrorMessage struct {
	Message  string
	CausedBy error
}

// MarshalJSON allows to marshal this object as a JSON object.
func (err ErrorMessage) MarshalJSON() ([]byte, error) {
	type JSONErrorMessage struct {
		Message string
		Cause   string
	}

	cause := ""
	if err.CausedBy != nil {
		cause = err.CausedBy.Error()
	}
	return json.Marshal(JSONErrorMessage{
		Message: err.Message,
		Cause:   cause,
	})
}

// String returns a string representation of the Error.
func (err ErrorMessage) String() string {
	if err.CausedBy == nil {
		return err.Message
	}
	return "Error: " + err.CausedBy.Error() + "\n" + err.Message
}

// PrintErrorMessage formats and prints info about an error message.
func PrintErrorMessage(msg string) {
	msg = strings.TrimSpace(msg)
	PrintError(nil, msg)
}

// PrintError formats and prints info about an error.
//
// Err is the error to print full info while msg is the user friendly message to print.
func PrintError(err error, msg string) {
	if logger != nil {
		logger.WithError(err).Error(msg)
	}
	Print(ErrorMessage{CausedBy: err, Message: msg})
}
