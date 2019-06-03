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
	IDEBuiltIn = iota
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
