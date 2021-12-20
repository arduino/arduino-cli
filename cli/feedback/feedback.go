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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/arduino/arduino-cli/i18n"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/status"
)

// OutputFormat is used to determine the output format
type OutputFormat int

const (
	// Text means plain text format, suitable for ansi terminals
	Text OutputFormat = iota
	// JSON means JSON format
	JSON
	// JSONMini is identical to JSON but without whitespaces
	JSONMini
	// YAML means YAML format
	YAML
)

// Result is anything more complex than a sentence that needs to be printed
// for the user.
type Result interface {
	fmt.Stringer
	Data() interface{}
}

// Feedback wraps an io.Writer and provides an uniform API the CLI can use to
// provide feedback to the users.
type Feedback struct {
	out    io.Writer
	err    io.Writer
	format OutputFormat
}

var tr = i18n.Tr

// New creates a Feedback instance
func New(out, err io.Writer, format OutputFormat) *Feedback {
	return &Feedback{
		out:    out,
		err:    err,
		format: format,
	}
}

// DefaultFeedback provides a basic feedback object to be used as default.
func DefaultFeedback() *Feedback {
	return New(os.Stdout, os.Stderr, Text)
}

// SetOut can be used to change the out writer at runtime
func (fb *Feedback) SetOut(out io.Writer) {
	fb.out = out
}

// SetErr can be used to change the err writer at runtime
func (fb *Feedback) SetErr(err io.Writer) {
	fb.err = err
}

// SetFormat can be used to change the output format at runtime
func (fb *Feedback) SetFormat(f OutputFormat) {
	fb.format = f
}

// GetFormat returns the output format currently set
func (fb *Feedback) GetFormat() OutputFormat {
	return fb.format
}

// OutputWriter returns the underlying io.Writer to be used when the Print*
// api is not enough.
func (fb *Feedback) OutputWriter() io.Writer {
	return fb.out
}

// ErrorWriter is the same as OutputWriter but exposes the underlying error
// writer.
func (fb *Feedback) ErrorWriter() io.Writer {
	return fb.out
}

// Printf behaves like fmt.Printf but writes on the out writer and adds a newline.
func (fb *Feedback) Printf(format string, v ...interface{}) {
	fb.Print(fmt.Sprintf(format, v...))
}

// Print behaves like fmt.Print but writes on the out writer and adds a newline.
func (fb *Feedback) Print(v interface{}) {
	switch fb.format {
	case JSON, JSONMini:
		fb.printJSON(v)
	case YAML:
		fb.printYAML(v)
	default:
		fmt.Fprintln(fb.out, v)
	}
}

// Errorf behaves like fmt.Printf but writes on the error writer and adds a
// newline. It also logs the error.
func (fb *Feedback) Errorf(format string, v ...interface{}) {
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
	fb.Error(fmt.Sprintf(format, v...))
}

// Error behaves like fmt.Print but writes on the error writer and adds a
// newline. It also logs the error.
func (fb *Feedback) Error(v ...interface{}) {
	fmt.Fprintln(fb.err, v...)
	logrus.Error(fmt.Sprint(v...))
}

// printJSON is a convenient wrapper to provide feedback by printing the
// desired output in a pretty JSON format. It adds a newline to the output.
func (fb *Feedback) printJSON(v interface{}) {
	var d []byte
	var err error
	if fb.format == JSON {
		d, err = json.MarshalIndent(v, "", "  ")
	} else if fb.format == JSONMini {
		d, err = json.Marshal(v)
	}
	if err != nil {
		fb.Errorf(tr("Error during JSON encoding of the output: %v"), err)
	} else {
		fmt.Fprintf(fb.out, "%v\n", string(d))
	}
}

// printYAML is a convenient wrapper to provide feedback by printing the
// desired output in YAML format. It adds a newline to the output.
func (fb *Feedback) printYAML(v interface{}) {
	d, err := yaml.Marshal(v)
	if err != nil {
		fb.Errorf(tr("Error during YAML encoding of the output: %v"), err)
		return
	}
	fmt.Fprintf(fb.out, "%v\n", string(d))
}

// PrintResult is a convenient wrapper to provide feedback for complex data,
// where the contents can't be just serialized to JSON but requires more
// structure.
func (fb *Feedback) PrintResult(res Result) {
	switch fb.format {
	case JSON, JSONMini:
		fb.printJSON(res.Data())
	case YAML:
		fb.printYAML(res.Data())
	default:
		fb.Print(fmt.Sprintf("%s", res))
	}
}
