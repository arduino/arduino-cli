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

// LibraryLocation represents where the library is installed
type LibraryLocation int

// The enumeration is listed in ascending order of priority
const (
	// IDEBuiltIn are libraries bundled in the IDE
	IDEBuiltIn LibraryLocation = iota
	// PlatformBuiltIn are libraries bundled in a PlatformRelease
	PlatformBuiltIn
	// ReferencedPlatformBuiltIn are libraries bundled in a PlatformRelease referenced for build
	ReferencedPlatformBuiltIn
	// User are user installed libraries
	User
)

func (d *LibraryLocation) String() string {
	switch *d {
	case IDEBuiltIn:
		return "ide"
	case PlatformBuiltIn:
		return "platform"
	case ReferencedPlatformBuiltIn:
		return "ref-platform"
	case User:
		return "user"
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
	case User:
		return json.Marshal("user")
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
	case "user":
		*d = User
	}
	return fmt.Errorf("invalid library location: %s", s)
}

// ToRPCLibraryLocation converts this LibraryLocation to rpc.LibraryLocation
func (d *LibraryLocation) ToRPCLibraryLocation() rpc.LibraryLocation {
	switch *d {
	case IDEBuiltIn:
		return rpc.LibraryLocation_ide_builtin
	case PlatformBuiltIn:
		return rpc.LibraryLocation_platform_builtin
	case ReferencedPlatformBuiltIn:
		return rpc.LibraryLocation_referenced_platform_builtin
	case User:
		return rpc.LibraryLocation_user
	}
	panic(fmt.Sprintf("invalid LibraryLocation value %d", *d))
}

// FromRPCLibraryLocation converts a rpc.LibraryLocation to a LibraryLocation
func FromRPCLibraryLocation(l rpc.LibraryLocation) LibraryLocation {
	switch l {
	case rpc.LibraryLocation_ide_builtin:
		return IDEBuiltIn
	case rpc.LibraryLocation_platform_builtin:
		return PlatformBuiltIn
	case rpc.LibraryLocation_referenced_platform_builtin:
		return ReferencedPlatformBuiltIn
	case rpc.LibraryLocation_user:
		return User
	}
	panic(fmt.Sprintf("invalid rpc.LibraryLocation value %d", l))
}
