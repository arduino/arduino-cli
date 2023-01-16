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

package monitor

import (
	"io"

	"github.com/arduino/arduino-cli/internal/cli/feedback"
)

type stdInOut struct {
	in  io.Reader
	out io.Writer
}

func newStdInOutTerminal() (*stdInOut, error) {
	in, out, err := feedback.InteractiveStreams()
	if err != nil {
		return nil, err
	}

	return &stdInOut{
		in:  in,
		out: out,
	}, nil
}

func (n *stdInOut) Close() error {
	return nil
}

func (n *stdInOut) Read(buff []byte) (int, error) {
	return n.in.Read(buff)
}

func (n *stdInOut) Write(buff []byte) (int, error) {
	return n.out.Write(buff)
}
