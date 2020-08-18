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

package executils

import "io"

// NullReader is an io.Reader that will always return EOF
var NullReader = &nullReader{}

type nullReader struct{}

func (r *nullReader) Read(buff []byte) (int, error) {
	return 0, io.EOF
}

// NullWriter is an io.Writer that discards any output
var NullWriter = &nullWriter{}

type nullWriter struct{}

func (r *nullWriter) Write(buff []byte) (int, error) {
	return len(buff), nil
}
