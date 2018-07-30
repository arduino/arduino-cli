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
