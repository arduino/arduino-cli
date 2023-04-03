// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
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

package f

// Filter takes a slice of type []T and a function of type func(T) bool. It returns
// a newly allocated slice containing only those elements of the input slice that
// satisfy the test function.
func Filter[T any](values []T, test func(x T) bool) []T {
	res := []T{}
	for _, x := range values {
		if test(x) {
			res = append(res, x)
		}
	}
	return res
}

// Map applies the mapping function to each element of the slice and returns a new
// slice with the results in the same order.
func Map[T any](values []T, mapping func(x T) T) []T {
	res := []T{}
	for _, x := range values {
		res = append(res, mapping(x))
	}
	return res
}

// Equals return a functor that matches the given value
func Equals[T comparable](value T) func(x T) bool {
	return func(x T) bool {
		return x == value
	}
}

// NotEquals return a functor that does not match the given value
func NotEquals[T comparable](value T) func(x T) bool {
	return func(x T) bool {
		return x != value
	}
}
