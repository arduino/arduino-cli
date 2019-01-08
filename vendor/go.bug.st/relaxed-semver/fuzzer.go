//
// Copyright 2018 Cristian Maglie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

// +build gofuzz

package semver

// Fuzz is used for fuzzy testing the project
func Fuzz(data []byte) int {
	v, err := Parse(string(data))
	if err != nil {
		if v != nil {
			panic("v != nil on error")
		}
		return 0
	}
	if v.String() != string(data) {
		panic("reserialized string != deserialized string")
	}
	v.Normalize()
	if v.CompareTo(v) != 0 {
		panic("compare != 0 while comparing with self")
	}
	r := ParseRelaxed(string(data))
	if r.String() != string(data) {
		panic("reserialized relaxed string != deserialized string")
	}
	return 1
}
