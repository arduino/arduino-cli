// This file is part of arduino-cli.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
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

package fqbn

import (
	"encoding/json"
	"fmt"
)

// UnmarshalJSON implements the json.Unmarshaler interface for the FQBN type.
func (f *FQBN) UnmarshalJSON(data []byte) error {
	var fqbnStr string
	if err := json.Unmarshal(data, &fqbnStr); err != nil {
		return fmt.Errorf("failed to unmarshal FQBN: %w", err)
	}

	fqbn, err := Parse(fqbnStr)
	if err != nil {
		return fmt.Errorf("invalid FQBN: %w", err)
	}

	*f = *fqbn
	return nil
}

// MarshalJSON implements the json.Marshaler interface for the FQBN type.
func (f FQBN) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}
