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
	stdOut io.Writer    = os.Stdout
	stdErr io.Writer    = os.Stderr
	format OutputFormat = Text
)

// Result is anything more complex than a sentence that needs to be printed
// for the user.
type Result interface {
	fmt.Stringer
	Data() interface{}
}

var tr = i18n.Tr

// SetOut can be used to change the out writer at runtime
func SetOut(out io.Writer) {
	stdOut = out
}

// SetErr can be used to change the err writer at runtime
func SetErr(err io.Writer) {
	stdErr = err
}

// SetFormat can be used to change the output format at runtime
func SetFormat(f OutputFormat) {
	format = f
}

// GetFormat returns the output format currently set
func GetFormat() OutputFormat {
	return format
}

// OutputWriter returns the underlying io.Writer to be used when the Print*
// api is not enough
func OutputWriter() io.Writer {
	return stdOut
}

// ErrorWriter is the same as OutputWriter but exposes the underlying error
// writer.
func ErrorWriter() io.Writer {
	return stdErr
}

// Printf behaves like fmt.Printf but writes on the out writer and adds a newline.
func Printf(format string, v ...interface{}) {
	Print(fmt.Sprintf(format, v...))
}

// Print behaves like fmt.Print but writes on the out writer and adds a newline.
func Print(v interface{}) {
	switch format {
	case JSON, MinifiedJSON:
		printJSON(v)
	case YAML:
		printYAML(v)
	case Text:
		fmt.Fprintln(stdOut, v)
	default:
		panic("unknown output format")
	}
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

// printJSON is a convenient wrapper to provide feedback by printing the
// desired output in a pretty JSON format. It adds a newline to the output.
func printJSON(v interface{}) {
	var d []byte
	var err error
	if format == JSON {
		d, err = json.MarshalIndent(v, "", "  ")
	} else if format == MinifiedJSON {
		d, err = json.Marshal(v)
	}
	if err != nil {
		Errorf(tr("Error during JSON encoding of the output: %v"), err)
	} else {
		fmt.Fprintf(stdOut, "%v\n", string(d))
	}
}

// printYAML is a convenient wrapper to provide feedback by printing the
// desired output in YAML format. It adds a newline to the output.
func printYAML(v interface{}) {
	d, err := yaml.Marshal(v)
	if err != nil {
		Errorf(tr("Error during YAML encoding of the output: %v"), err)
		return
	}
	fmt.Fprintf(stdOut, "%v\n", string(d))
}

// PrintResult is a convenient wrapper to provide feedback for complex data,
// where the contents can't be just serialized to JSON but requires more
// structure.
func PrintResult(res Result) {
	switch format {
	case JSON, MinifiedJSON:
		printJSON(res.Data())
	case YAML:
		printYAML(res.Data())
	case Text:
		Print(res.String())
	default:
		panic("unknown output format")
	}
}
