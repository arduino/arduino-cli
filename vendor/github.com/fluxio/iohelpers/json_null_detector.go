package iohelpers

import (
	"io"
	"unicode"
)

const (
	stateInit    = iota
	stateN       = iota
	stateU       = iota
	stateL       = iota
	stateAccept  = iota
	stateInvalid = iota
)

// JsonNullDetector is an io.Reader that passes through Read() results from an
// underlying source reader, while detecting whether those bytes constitute a
// valid representation of JSON null, i.e. the four bytes "null" surrounded by
// JSON whitespace.  Note that although JSON is UTF-8, its definition of
// syntactic whitespace only includes space, tab, newline, and carriage return.
// This can be done in a streaming fashion with a finite state machine.
type JsonNullDetector struct {
	io.Reader
	state int
}

func (d *JsonNullDetector) Read(p []byte) (int, error) {
	n, err := d.Reader.Read(p)

	if n == 0 || d.state == stateInvalid {
		return n, err
	}

	for i := 0; i < n; i++ {
		b := p[i]
		if b > unicode.MaxASCII {
			d.state = stateInvalid
		} else {
			d.state = d.nextState(b)
		}
		if d.state == stateInvalid {
			break
		}
	}

	return n, err
}

// Detected returns true if JSON null, and nothing else besides leading/trailing
// whitespace, was detected on the prefix of the stream read so far.
func (d *JsonNullDetector) Detected() bool {
	return d.state == stateAccept
}

// nextState returns the next state for d's state machine after consuming b.
func (d *JsonNullDetector) nextState(b byte) int {
	switch d.state {
	case stateInit:
		if isJsonSpace(b) {
			return stateInit
		}
		if b == 'n' {
			return stateN
		}
	case stateN:
		if b == 'u' {
			return stateU
		}
	case stateU:
		if b == 'l' {
			return stateL
		}
	case stateL:
		if b == 'l' {
			return stateAccept
		}
	case stateAccept:
		if isJsonSpace(b) {
			return stateAccept
		}
	}
	return stateInvalid
}

func isJsonSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}
