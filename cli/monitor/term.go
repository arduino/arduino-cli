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
	"os"

	"github.com/mattn/go-tty"
)

type nullTerminal struct {
}

func newNullTerminal() (*nullTerminal, error) {
	return &nullTerminal{}, nil
}

func (n *nullTerminal) Close() error {
	return nil
}

func (n *nullTerminal) Read(buff []byte) (int, error) {
	return os.Stdin.Read(buff)
}

func (n *nullTerminal) Write(buff []byte) (int, error) {
	return os.Stdout.Write(buff)
}

func (n *nullTerminal) AddEscapeCallback(func(r rune) bool) {
}

type terminal struct {
	clean  func() error
	term   *tty.TTY
	escape func(rune) bool
}

func newTerminal() (*terminal, error) {
	term, err := tty.Open()
	if err != nil {
		return nil, err
		// } else if clean, err := term.Raw(); err != nil {
		// 	return nil, err
	}
	return &terminal{
		// clean: clean,
		term: term,
	}, nil
}

func (t *terminal) Close() error {
	if t.clean != nil {
		return t.clean()
	}
	return nil
}

func (t *terminal) Read(buff []byte) (int, error) {
	for {
		r, err := t.term.ReadRune()
		if err != nil {
			return 0, err
		}
		if r == 27 && t.escape != nil {
			for t.escape(r) {
				// continue to feed chars to escape handler if requested..-
				r, err = t.term.ReadRune()
				if err != nil {
					return 0, err
				}
			}
		} else {
			return copy(buff, []byte(string(r))), nil
		}
	}
}

func (t *terminal) Write(buff []byte) (int, error) {
	return os.Stdout.Write(buff)
}

func (t *terminal) AddEscapeCallback(cb func(rune) bool) {
	t.escape = cb
}
