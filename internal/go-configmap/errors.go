// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
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

package configmap

import "strings"

// UnmarshalErrors is a collection of errors that occurred during unmarshalling.
// Do not return this type directly, but use its result() method instead.
type UnmarshalErrors struct {
	wrapped []error
}

func (e *UnmarshalErrors) append(err error) {
	e.wrapped = append(e.wrapped, err)
}

func (e *UnmarshalErrors) result() error {
	if len(e.wrapped) == 0 {
		return nil
	}
	return e
}

func (e *UnmarshalErrors) Error() string {
	if len(e.wrapped) == 1 {
		return e.wrapped[0].Error()
	}
	errs := []string{"multiple errors:"}
	for _, err := range e.wrapped {
		errs = append(errs, "- "+err.Error())
	}
	return strings.Join(errs, "\n")
}

// WrappedErrors returns the list of errors that occurred during unmarshalling.
func (e *UnmarshalErrors) WrappedErrors() []error {
	if e == nil {
		return nil
	}
	return e.wrapped
}
