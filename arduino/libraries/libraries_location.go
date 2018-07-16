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

// LibraryLocation represents where the library is installed
type LibraryLocation int

// The enumeration is listed in ascending order of priority
const (
	// IDEBuiltIn are libraries bundled in the IDE
	IDEBuiltIn = 1 << iota
	// PlatformBuiltIn are libraries bundled in a PlatformRelease
	PlatformBuiltIn
	// ReferencedPlatformBuiltIn are libraries bundled in a PlatformRelease referenced for build
	ReferencedPlatformBuiltIn
	// Sketchbook are user installed libraries
	Sketchbook
)

func (d *LibraryLocation) String() string {
	switch *d {
	case IDEBuiltIn:
		return "ide"
	case PlatformBuiltIn:
		return "platform"
	case ReferencedPlatformBuiltIn:
		return "ref-platform"
	case Sketchbook:
		return "sketchbook"
	}
	panic(fmt.Sprintf("invalid LibraryLocation value %d", *d))
}

// MarshalJSON implements the json.Marshaler interface
func (d *LibraryLocation) MarshalJSON() ([]byte, error) {
	switch *d {
	case IDEBuiltIn:
		return json.Marshal("ide")
	case PlatformBuiltIn:
		return json.Marshal("platform")
	case ReferencedPlatformBuiltIn:
		return json.Marshal("ref-platform")
	case Sketchbook:
		return json.Marshal("sketchbook")
	}
	return nil, fmt.Errorf("invalid library location value: %d", *d)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (d *LibraryLocation) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch s {
	case "ide":
		*d = IDEBuiltIn
	case "platform":
		*d = PlatformBuiltIn
	case "ref-platform":
		*d = ReferencedPlatformBuiltIn
	case "sketchbook":
		*d = Sketchbook
	}
	return fmt.Errorf("invalid library location: %s", s)
}
