//
// Copyright 2018 Cristian Maglie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package semver

import (
	"encoding/json"
)

// MarshalJSON implements json.Marshaler
func (v *Version) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

// UnmarshalJSON implements json.Unmarshaler
func (v *Version) UnmarshalJSON(data []byte) error {
	var versionString string
	if err := json.Unmarshal(data, &versionString); err != nil {
		return err
	}
	parsed, err := Parse(versionString)
	if err != nil {
		return err
	}

	v.major = parsed.major
	v.minor = parsed.minor
	v.patch = parsed.patch
	v.prerelases = parsed.prerelases
	v.numericPrereleases = parsed.numericPrereleases
	v.builds = parsed.builds
	return nil
}

// MarshalJSON implements json.Marshaler
func (v *RelaxedVersion) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

// UnmarshalJSON implements json.Unmarshaler
func (v *RelaxedVersion) UnmarshalJSON(data []byte) error {
	var versionString string
	if err := json.Unmarshal(data, &versionString); err != nil {
		return err
	}
	parsed := ParseRelaxed(versionString)

	v.customversion = parsed.customversion
	v.version = parsed.version
	return nil
}
