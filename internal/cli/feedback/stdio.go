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

// OutputStreams returns the underlying io.Writer to directly stream to
// stdout and stderr.
// If the selected output format is not Text, the returned writers will
// accumulate the output until command execution is completed.
// This function returns also a callback that must be called when the
// command execution is completed, it will return a *OutputStreamsResult
// object that can be used as a Result or to retrieve the output to embed
// it in another object.
func OutputStreams() (io.Writer, io.Writer, func() *OutputStreamsResult) {
	if !formatSelected {
		panic("output format not yet selected")
	}
	return feedbackOut, feedbackErr, getOutputStreamResult
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
	// In case of non-Text output format, the output is accumulared so retrun the buffer as a Result object
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
