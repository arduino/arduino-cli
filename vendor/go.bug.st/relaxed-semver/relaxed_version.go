//
// Copyright 2018 Cristian Maglie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package semver

import "fmt"

// RelaxedVersion allows any possible version string. If the version does not comply
// with semantic versioning it is saved as-is and only Equal comparison will match.
type RelaxedVersion struct {
	customversion []byte
	version       *Version
}

// WarnInvalidVersionWhenParsingRelaxed must be set to true to show warnings while
// parsing RelaxedVersion if an invalid semver string is found. This allows a soft
// transition to strict semver
var WarnInvalidVersionWhenParsingRelaxed = false

// ParseRelaxed parse a RelaxedVersion
func ParseRelaxed(in string) *RelaxedVersion {
	v, err := Parse(in)
	if err == nil {
		return &RelaxedVersion{version: v}
	}
	if WarnInvalidVersionWhenParsingRelaxed {
		fmt.Printf("WARNING invalid semver version %s: %s\n", in, err)
	}
	return &RelaxedVersion{customversion: []byte(in[:])}
}

func (v *RelaxedVersion) String() string {
	if v == nil {
		return ""
	}
	if v.version != nil {
		return v.version.String()
	}
	return string(v.customversion)
}

// CompareTo compares the RelaxedVersion with the one passed as parameter.
// Returns -1, 0 or 1 if the version is respectively less than, equal
// or greater than the compared Version
func (v *RelaxedVersion) CompareTo(u *RelaxedVersion) int {
	if v.version == nil && u.version == nil {
		return compareAlpha(v.customversion, u.customversion)
	}
	if v.version == nil {
		return -1
	}
	if u.version == nil {
		return 1
	}
	return v.version.CompareTo(u.version)
}

// LessThan returns true if the RelaxedVersion is less than the RelaxedVersion passed as parameter
func (v *RelaxedVersion) LessThan(u *RelaxedVersion) bool {
	return v.CompareTo(u) < 0
}

// LessThanOrEqual returns true if the RelaxedVersion is less than or equal to the RelaxedVersion passed as parameter
func (v *RelaxedVersion) LessThanOrEqual(u *RelaxedVersion) bool {
	return v.CompareTo(u) <= 0
}

// Equal returns true if the RelaxedVersion is equal to the RelaxedVersion passed as parameter
func (v *RelaxedVersion) Equal(u *RelaxedVersion) bool {
	return v.CompareTo(u) == 0
}

// GreaterThan returns true if the RelaxedVersion is greater than the RelaxedVersion passed as parameter
func (v *RelaxedVersion) GreaterThan(u *RelaxedVersion) bool {
	return v.CompareTo(u) > 0
}

// GreaterThanOrEqual returns true if the RelaxedVersion is greater than or equal to the RelaxedVersion passed as parameter
func (v *RelaxedVersion) GreaterThanOrEqual(u *RelaxedVersion) bool {
	return v.CompareTo(u) >= 0
}
