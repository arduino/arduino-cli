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

package fqbn

import (
	"errors"
	"regexp"
	"strings"

	"github.com/arduino/arduino-cli/internal/i18n"
	properties "github.com/arduino/go-properties-orderedmap"
)

// FQBN represents an Fully Qualified Board Name string
type FQBN struct {
	Packager     string
	Architecture string
	BoardID      string
	Configs      *properties.Map
}

// MustParse parse an FQBN string from the input string
// or panics if the input is not a valid FQBN.
func MustParse(fqbnIn string) *FQBN {
	res, err := Parse(fqbnIn)
	if err != nil {
		panic(err)
	}
	return res
}

var fqbnValidationRegex = regexp.MustCompile(`^[a-zA-Z0-9_.-]*$`)
var valueValidationRegex = regexp.MustCompile(`^[a-zA-Z0-9=_.-]*$`)

// Parse parses an FQBN string from the input string
func Parse(fqbnIn string) (*FQBN, error) {
	// Split fqbn parts
	fqbnParts := strings.Split(fqbnIn, ":")
	if len(fqbnParts) < 3 || len(fqbnParts) > 4 {
		return nil, errors.New(i18n.Tr("not an FQBN: %s", fqbnIn))
	}

	fqbn := &FQBN{
		Packager:     fqbnParts[0],
		Architecture: fqbnParts[1],
		BoardID:      fqbnParts[2],
		Configs:      properties.NewMap(),
	}
	if fqbn.BoardID == "" {
		return nil, errors.New(i18n.Tr("empty board identifier"))
	}
	// Check if the fqbn contains invalid characters
	for i := 0; i < 3; i++ {
		if !fqbnValidationRegex.MatchString(fqbnParts[i]) {
			return nil, errors.New(i18n.Tr("fqbn's field %s contains an invalid character", fqbnParts[i]))
		}
	}
	if len(fqbnParts) > 3 {
		for _, pair := range strings.Split(fqbnParts[3], ",") {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) != 2 {
				return nil, errors.New(i18n.Tr("invalid config option: %s", pair))
			}
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			if k == "" {
				return nil, errors.New(i18n.Tr("invalid config option: %s", pair))
			}
			if !fqbnValidationRegex.MatchString(k) {
				return nil, errors.New(i18n.Tr("config key %s contains an invalid character", k))
			}
			// The config value can also contain the = symbol
			if !valueValidationRegex.MatchString(v) {
				return nil, errors.New(i18n.Tr("config value %s contains an invalid character", v))
			}
			fqbn.Configs.Set(k, v)
		}
	}
	return fqbn, nil
}

// Clone returns a copy of this FQBN.
func (fqbn *FQBN) Clone() *FQBN {
	return &FQBN{
		Packager:     fqbn.Packager,
		Architecture: fqbn.Architecture,
		BoardID:      fqbn.BoardID,
		Configs:      fqbn.Configs.Clone(),
	}
}

// Match checks if the target FQBN equals to this one.
// The core parts are checked for exact equality while board options are loosely
// matched: the set of boards options of the target must be fully contained within
// the one of the receiver and their values must be equal.
func (fqbn *FQBN) Match(target *FQBN) bool {
	if fqbn.StringWithoutConfig() != target.StringWithoutConfig() {
		return false
	}

	for neededKey, neededValue := range target.Configs.AsMap() {
		targetValue, hasKey := fqbn.Configs.GetOk(neededKey)
		if !hasKey || targetValue != neededValue {
			return false
		}
	}
	return true
}

// StringWithoutConfig returns the FQBN without the Config part
func (fqbn *FQBN) StringWithoutConfig() string {
	return fqbn.Packager + ":" + fqbn.Architecture + ":" + fqbn.BoardID
}

// String returns the FQBN as a string
func (fqbn *FQBN) String() string {
	res := fqbn.StringWithoutConfig()
	if fqbn.Configs.Size() > 0 {
		sep := ":"
		for _, k := range fqbn.Configs.Keys() {
			res += sep + k + "=" + fqbn.Configs.Get(k)
			sep = ","
		}
	}
	return res
}
