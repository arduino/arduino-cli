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
	"bytes"
	"errors"
	"io"
)

// DirectStreams returns the underlying io.Writer to directly stream to
// stdout and stderr.
// If the selected output format is not Text, the function will error.
//
// Using the streams returned by this function allows direct control of
// the output and the PrintResult function must not be used anymore
func DirectStreams() (io.Writer, io.Writer, error) {
	if !formatSelected {
		panic("output format not yet selected")
	}
	if format != Text {
		return nil, nil, errors.New(tr("available only in text format"))
	}
	return stdOut, stdErr, nil
}

// OutputStreams returns a pair of io.Writer to write the command output.
// The returned writers will accumulate the output until the command
// execution is completed, so they are not suitable for printing an unbounded
// stream like a debug logger or an event watcher (use DirectStreams for
// that purpose).
//
// If the output format is Text the output will be directly streamed to the
// underlying stdio streams in real time.
//
// This function returns also a callback that must be called when the
// command execution is completed, it will return an *OutputStreamsResult
// object that can be used as a Result or to retrieve the accumulated output
// to embed it in another object.
func OutputStreams() (io.Writer, io.Writer, func() *OutputStreamsResult) {
	if !formatSelected {
		panic("output format not yet selected")
	}
	return feedbackOut, feedbackErr, getOutputStreamResult
}

// NewBufferedStreams returns a pair of io.Writer to buffer the command output.
// The returned writers will accumulate the output until the command
// execution is completed. The io.Writes will not affect other feedback streams.
//
// This function returns also a callback that must be called when the
// command execution is completed, it will return an *OutputStreamsResult
// object that can be used as a Result or to retrieve the accumulated output
// to embed it in another object.
func NewBufferedStreams() (io.Writer, io.Writer, func() *OutputStreamsResult) {
	out, err := &bytes.Buffer{}, &bytes.Buffer{}
	return out, err, func() *OutputStreamsResult {
		return &OutputStreamsResult{
			Stdout: out.String(),
			Stderr: err.String(),
		}
	}
}

func getOutputStreamResult() *OutputStreamsResult {
	return &OutputStreamsResult{
		Stdout: bufferOut.String(),
		Stderr: bufferErr.String(),
	}
}

// OutputStreamsResult contains the accumulated stdout and stderr output
// when the selected output format is not Text.
type OutputStreamsResult struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

// Data returns the result object itself, it is used to implement the Result interface.
func (r *OutputStreamsResult) Data() interface{} {
	// In case of non-Text output format, the output is accumulated so return the buffer as a Result object
	return r
}

func (r *OutputStreamsResult) String() string {
	// In case of Text output format, the output is streamed to stdout and stderr directly, no need to print anything
	return ""
}

// Empty returns true if both Stdout and Stderr are empty.
func (r *OutputStreamsResult) Empty() bool {
	return r.Stdout == "" && r.Stderr == ""
}
