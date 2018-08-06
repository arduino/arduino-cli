//
// Copyright 2018 Cristian Maglie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package semver

// Version contains the results of parsed version string
type Version struct {
	major              []byte
	minor              []byte
	patch              []byte
	prerelases         [][]byte
	numericPrereleases []bool
	builds             [][]byte
}

func (v *Version) String() string {
	res := string(v.major)
	if len(v.minor) > 0 {
		res += "." + string(v.minor)
	}
	if len(v.patch) > 0 {
		res += "." + string(v.patch)
	}
	for i, prerelease := range v.prerelases {
		if i == 0 {
			res += "-"
		} else {
			res += "."
		}
		res += string(prerelease)
	}
	for i, build := range v.builds {
		if i == 0 {
			res += "+"
		} else {
			res += "."
		}
		res += string(build)
	}
	return res
}

var zero = []byte("0")

// Normalize transforms a truncated semver version in a strictly compliant semver
// version by adding minor and patch versions. For example:
// "1" is trasformed to "1.0.0" or "2.5-dev" to "2.5.0-dev"
func (v *Version) Normalize() {
	if len(v.major) == 0 {
		v.major = zero[0:1]
	}
	if len(v.minor) == 0 {
		v.minor = zero[0:1]
	}
	if len(v.patch) == 0 {
		v.patch = zero[0:1]
	}
}

func compareNumber(a, b []byte) int {
	la := len(a)
	if la == 0 {
		a = zero[:]
		la = 1
	}
	lb := len(b)
	if lb == 0 {
		b = zero[:]
		lb = 1
	}
	if la == lb {
		for i := range a {
			if a[i] > b[i] {
				return 1
			}
			if a[i] < b[i] {
				return -1
			}
		}
		return 0
	}
	if la > lb {
		return 1
	}
	return -1
}

func compareAlpha(a, b []byte) int {
	if string(a) > string(b) {
		return 1
	}
	if string(a) < string(b) {
		return -1
	}
	return 0
}

// CompareTo compares the Version with the one passed as parameter.
// Returns -1, 0 or 1 if the version is respectively less than, equal
// or greater than the compared Version
func (v *Version) CompareTo(u *Version) int {
	// 11. Precedence refers to how versions are compared to each other when ordered.
	// Precedence MUST be calculated by separating the version into major, minor,
	// patch and pre-release identifiers in that order (Build metadata does not
	// figure into precedence). Precedence is determined by the first difference when
	// comparing each of these identifiers from left to right as follows: Major, minor,
	// and patch versions are always compared numerically.
	// Example: 1.0.0 < 2.0.0 < 2.1.0 < 2.1.1.
	major := compareNumber(v.major, u.major)
	if major != 0 {
		return major
	}
	minor := compareNumber(v.minor, u.minor)
	if minor != 0 {
		return minor
	}
	patch := compareNumber(v.patch, u.patch)
	if patch != 0 {
		return patch
	}

	// When major, minor, and patch are equal, a pre-release version has lower
	// precedence than a normal version.
	// Example: 1.0.0-alpha < 1.0.0.
	lv := len(v.prerelases)
	lu := len(u.prerelases)
	if lv == 0 && lu == 0 {
		return 0
	}
	if lv == 0 {
		return 1
	}
	if lu == 0 {
		return -1
	}

	// Precedence for two pre-release versions with the same major, minor, and patch
	// version MUST be determined by comparing each dot separated identifier from left
	// to right until a difference is found as follows:
	// - identifiers consisting of only digits are compared numerically
	// - identifiers with letters or hyphens are compared lexically in ASCII sort order.
	// Numeric identifiers always have lower precedence than non-numeric identifiers.
	// A larger set of pre-release fields has a higher precedence than a smaller set,
	// if all of the preceding identifiers are equal.
	// Example: 1.0.0-alpha < 1.0.0-alpha.1 < 1.0.0-alpha.beta < 1.0.0-beta <
	//          < 1.0.0-beta.2 < 1.0.0-beta.11 < 1.0.0-rc.1 < 1.0.0.
	min := lv
	if lv > lu {
		min = lu
	}
	for i := 0; i < min; i++ {
		if v.numericPrereleases[i] && u.numericPrereleases[i] {
			comp := compareNumber(v.prerelases[i], u.prerelases[i])
			if comp != 0 {
				return comp
			}
			continue
		}
		if v.numericPrereleases[i] {
			return -1
		}
		if u.numericPrereleases[i] {
			return 1
		}
		comp := compareAlpha(v.prerelases[i], u.prerelases[i])
		if comp != 0 {
			return comp
		}
	}
	if lv > lu {
		return 1
	}
	if lv < lu {
		return -1
	}
	return 0
}

// LessThan returns true if the Version is less than the Version passed as parameter
func (v *Version) LessThan(u *Version) bool {
	return v.CompareTo(u) < 0
}

// LessThanOrEqual returns true if the Version is less than or equal to the Version passed as parameter
func (v *Version) LessThanOrEqual(u *Version) bool {
	return v.CompareTo(u) <= 0
}

// Equal returns true if the Version is equal to the Version passed as parameter
func (v *Version) Equal(u *Version) bool {
	return v.CompareTo(u) == 0
}

// GreaterThan returns true if the Version is greater than the Version passed as parameter
func (v *Version) GreaterThan(u *Version) bool {
	return v.CompareTo(u) > 0
}

// GreaterThanOrEqual returns true if the Version is greater than or equal to the Version passed as parameter
func (v *Version) GreaterThanOrEqual(u *Version) bool {
	return v.CompareTo(u) >= 0
}
