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

// Package feedback provides an uniform API that can be used to
// print feedback to the users in different formats.
package feedback

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/arduino/arduino-cli/i18n"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v2"
)

// OutputFormat is an output format
type OutputFormat int

const (
	// Text is the plain text format, suitable for interactive terminals
	Text OutputFormat = iota
	// JSON format
	JSON
	// MinifiedJSON format
	MinifiedJSON
	// YAML format
	YAML
)

var formats map[string]OutputFormat = map[string]OutputFormat{
	"json":     JSON,
	"jsonmini": MinifiedJSON,
	"yaml":     YAML,
	"text":     Text,
}

func (f OutputFormat) String() string {
	for res, format := range formats {
		if format == f {
			return res
		}
	}
	panic("unknown output format")
}

// ParseOutputFormat parses a string and returns the corresponding OutputFormat.
// The boolean returned is true if the string was a valid OutputFormat.
func ParseOutputFormat(in string) (OutputFormat, bool) {
	format, found := formats[in]
	return format, found
}

var (
	stdOut         io.Writer
	stdErr         io.Writer
	feedbackOut    io.Writer
	feedbackErr    io.Writer
	bufferOut      *bytes.Buffer
	bufferErr      *bytes.Buffer
	format         OutputFormat
	formatSelected bool
)

func init() {
	reset()
}

// reset resets the feedback package to its initial state, useful for unit testing
func reset() {
	stdOut = os.Stdout
	stdErr = os.Stderr
	feedbackOut = os.Stdout
	feedbackErr = os.Stderr
	bufferOut = &bytes.Buffer{}
	bufferErr = &bytes.Buffer{}
	format = Text
	formatSelected = false
}

// Result is anything more complex than a sentence that needs to be printed
// for the user.
type Result interface {
	fmt.Stringer
	Data() interface{}
}

var tr = i18n.Tr

// SetOut can be used to change the out writer at runtime
func SetOut(out io.Writer) {
	if formatSelected {
		panic("output format already selected")
	}
	stdOut = out
}

// SetErr can be used to change the err writer at runtime
func SetErr(err io.Writer) {
	if formatSelected {
		panic("output format already selected")
	}
	stdErr = err
}

// SetFormat can be used to change the output format at runtime
func SetFormat(f OutputFormat) {
	if formatSelected {
		panic("output format already selected")
	}
	format = f
	formatSelected = true

	if format == Text {
		feedbackOut = stdOut
		feedbackErr = stdErr
	} else {
		feedbackOut = bufferOut
		feedbackErr = bufferErr
	}
}

// GetFormat returns the output format currently set
func GetFormat() OutputFormat {
	return format
}

// Printf behaves like fmt.Printf but writes on the out writer and adds a newline.
func Printf(format string, v ...interface{}) {
	Print(fmt.Sprintf(format, v...))
}

// Print behaves like fmt.Print but writes on the out writer and adds a newline.
func Print(v string) {
	fmt.Fprintln(feedbackOut, v)
}

// Errorf behaves like fmt.Printf but writes on the error writer and adds a
// newline. It also logs the error.
func Errorf(format string, v ...interface{}) {
	// Unbox grpc status errors
	for i := range v {
		if s, isStatus := v[i].(*status.Status); isStatus {
			v[i] = errors.New(s.Message())
		} else if err, isErr := v[i].(error); isErr {
			if s, isStatus := status.FromError(err); isStatus {
				v[i] = errors.New(s.Message())
			}
		}
	}
	Error(fmt.Sprintf(format, v...))
}

// Error behaves like fmt.Print but writes on the error writer and adds a
// newline. It also logs the error.
func Error(v ...interface{}) {
	fmt.Fprintln(stdErr, v...)
	logrus.Error(fmt.Sprint(v...))
}

// PrintResult is a convenient wrapper to provide feedback for complex data,
// where the contents can't be just serialized to JSON but requires more
// structure.
func PrintResult(res Result) {
	var data string
	switch format {
	case JSON:
		d, err := json.MarshalIndent(res.Data(), "", "  ")
		if err != nil {
			Errorf("Error during JSON encoding of the output: %v", err)
			return
		}
		data = string(d)
	case MinifiedJSON:
		d, err := json.Marshal(res.Data())
		if err != nil {
			Errorf("Error during JSON encoding of the output: %v", err)
			return
		}
		data = string(d)
	case YAML:
		d, err := yaml.Marshal(res.Data())
		if err != nil {
			Errorf("Error during YAML encoding of the output: %v", err)
			return
		}
		data = string(d)
	case Text:
		data = res.String()
	default:
		panic("unknown output format")
	}
	if data != "" {
		fmt.Fprintln(stdOut, data)
	}
}
