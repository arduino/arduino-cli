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

// Matcher is a function that tests if a given value match a certain criteria.
type Matcher[T any] func(T) bool

// Reducer is a function that combines two values of the same type and return
// the combined value.
type Reducer[T any] func(T, T) T

// Mapper is a function that converts a value of one type to another type.
type Mapper[T, U any] func(T) U

// Filter takes a slice of type []T and a Matcher[T]. It returns a newly
// allocated slice containing only those elements of the input slice that
// satisfy the matcher.
func Filter[T any](values []T, matcher Matcher[T]) []T {
	res := []T{}
	for _, x := range values {
		if matcher(x) {
			res = append(res, x)
		}
	}
	return res
}

// Map applies the Mapper function to each element of the slice and returns
// a new slice with the results in the same order.
func Map[T, U any](values []T, mapper Mapper[T, U]) []U {
	res := []U{}
	for _, x := range values {
		res = append(res, mapper(x))
	}
	return res
}

// Reduce applies the Reducer function to all elements of the input values
// and returns the result.
func Reduce[T any](values []T, reducer Reducer[T]) T {
	var result T
	for _, v := range values {
		result = reducer(result, v)
	}
	return result
}

// Equals return a Matcher that matches the given value
func Equals[T comparable](value T) Matcher[T] {
	return func(x T) bool {
		return x == value
	}
}

// NotEquals return a Matcher that does not match the given value
func NotEquals[T comparable](value T) Matcher[T] {
	return func(x T) bool {
		return x != value
	}
}
