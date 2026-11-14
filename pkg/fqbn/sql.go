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

import "fmt"

// Value implements the driver.Valuer interface for the FQBN type.
func (f FQBN) Value() (any, error) {
	return f.String(), nil
}

// Scan implements the sql.Scanner interface for the FQBN type.
func (f *FQBN) Scan(value any) error {
	if value == nil {
		return nil
	}

	if v, ok := value.(string); ok {
		ParsedFQBN, err := Parse(v)
		if err != nil {
			return fmt.Errorf("failed to parse FQBN: %w", err)
		}
		*f = *ParsedFQBN
		return nil
	}

	return fmt.Errorf("unsupported type: %T", value)
}
