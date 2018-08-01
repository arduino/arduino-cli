/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO AG (http://www.arduino.cc/)
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
 */

package libraries

import (
	"encoding/json"
	"fmt"
)

// LibraryLayout represents how the library source code is layed out in the library
type LibraryLayout uint16

const (
	// FlatLayout is a library without a `src` directory
	FlatLayout LibraryLayout = 1 << iota
	// RecursiveLayout is a library with `src` directory (that allows recursive build)
	RecursiveLayout
)

func (d *LibraryLayout) String() string {
	switch *d {
	case FlatLayout:
		return "flat"
	case RecursiveLayout:
		return "recursive"
	}
	panic(fmt.Sprintf("invalid LibraryLayout value %d", *d))
}

// MarshalJSON implements the json.Marshaler interface
func (d *LibraryLayout) MarshalJSON() ([]byte, error) {
	switch *d {
	case FlatLayout:
		return json.Marshal("flat")
	case RecursiveLayout:
		return json.Marshal("recursive")
	}
	return nil, fmt.Errorf("invalid library layout value: %d", *d)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (d *LibraryLayout) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch s {
	case "flat":
		*d = FlatLayout
	case "recursive":
		*d = RecursiveLayout
	}
	return fmt.Errorf("invalid library layout: %s", s)
}
