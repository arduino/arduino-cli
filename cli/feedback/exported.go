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

package feedback

import (
	"io"
)

var (
	fb = DefaultFeedback()
)

// SetDefaultFeedback lets callers override the default feedback object. Mostly
// useful for testing.
func SetDefaultFeedback(f *Feedback) {
	fb = f
}

// SetFormat can be used to change the output format at runtime
func SetFormat(f OutputFormat) {
	fb.SetFormat(f)
}

// GetFormat returns the currently set output format
func GetFormat() OutputFormat {
	return fb.GetFormat()
}

// OutputWriter returns the underlying io.Writer to be used when the Print*
// api is not enough
func OutputWriter() io.Writer {
	return fb.OutputWriter()
}

// ErrorWriter is the same as OutputWriter but exposes the underlying error
// writer
func ErrorWriter() io.Writer {
	return fb.ErrorWriter()
}

// Printf behaves like fmt.Printf but writes on the out writer and adds a newline.
func Printf(format string, v ...interface{}) {
	fb.Printf(format, v...)
}

// Print behaves like fmt.Print but writes on the out writer and adds a newline.
func Print(v interface{}) {
	fb.Print(v)
}

// Errorf behaves like fmt.Printf but writes on the error writer and adds a
// newline. It also logs the error.
func Errorf(format string, v ...interface{}) {
	fb.Errorf(format, v...)
}

// Error behaves like fmt.Print but writes on the error writer and adds a
// newline. It also logs the error.
func Error(v ...interface{}) {
	fb.Error(v...)
}

// PrintResult is a convenient wrapper to provide feedback for complex data,
// where the contents can't be just serialized to JSON but requires more
// structure.
func PrintResult(res Result) {
	fb.PrintResult(res)
}
