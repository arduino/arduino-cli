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
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
 */

package formatter

import (
	"encoding/json"
	"fmt"
)

// ErrorMessage represents an Error with an attached message.
//
// It's the same as a normal error, but It is also parsable as JSON.
type ErrorMessage struct {
	message string
}

// MarshalJSON allows to marshal this object as a JSON object.
func (err ErrorMessage) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.message)
}

// Error returns the error message.
func (err ErrorMessage) Error() string {
	return fmt.Sprint(err.message)
}

// String returns a string representation of the Error.
func (err ErrorMessage) String() string {
	return err.Error()
}

// romError creates an ErrorMessage from an Error.
func fromError(err error) ErrorMessage {
	return ErrorMessage{
		message: err.Error(),
	}
}
