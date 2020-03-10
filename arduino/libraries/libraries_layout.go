// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package libraries

import (
	"encoding/json"
	"fmt"

	rpc "github.com/arduino/arduino-cli/rpc/commands"
)

// LibraryLayout represents how the library source code is laid out in the library
type LibraryLayout uint16

const (
	// FlatLayout is a library without a `src` directory
	FlatLayout LibraryLayout = iota
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

// ToRPCLibraryLayout converts this LibraryLayout to rpc.LibraryLayout
func (d *LibraryLayout) ToRPCLibraryLayout() rpc.LibraryLayout {
	switch *d {
	case FlatLayout:
		return rpc.LibraryLayout_flat_layout
	case RecursiveLayout:
		return rpc.LibraryLayout_recursive_layout
	}
	panic(fmt.Sprintf("invalid LibraryLayout value %d", *d))
}
