//
// Copyright 2018 Cristian Maglie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package semver

/*
goos: linux
goarch: amd64
pkg: go.bug.st/semver
BenchmarkNumericArray-8               	2000000000	         1.71 ns/op
BenchmarkNumericFunction-8            	2000000000	         1.23 ns/op <BEST
BenchmarkIdentifierArray-8            	1000000000	         2.23 ns/op <BEST
BenchmarkIdentifierFunction-8         	300000000	         5.92 ns/op
BenchmarkVersionSeparatorArray-8      	2000000000	         1.70 ns/op <BEST
BenchmarkVersionSeparatorFunction-8   	300000000	         5.10 ns/op
*/

func isNumeric(c byte) bool {
	return c >= '0' && c <= '9'
}

func isIdentifier(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '-'
}

func isVersionSeparator(c byte) bool {
	return c == '.' || c == '-' || c == '+'
}

var numeric [256]bool
var identifier [256]bool
var versionSeparator [256]bool

func init() {
	for i := 0; i < 256; i++ {
		c := byte(i)
		if isNumeric(c) {
			numeric[c] = true
		}
		if isIdentifier(c) {
			identifier[c] = true
		}
		if isVersionSeparator(c) {
			versionSeparator[c] = true
		}
	}
}
