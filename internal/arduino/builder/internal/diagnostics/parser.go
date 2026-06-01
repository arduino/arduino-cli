// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
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

package diagnostics

import (
	"fmt"
	"strings"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// CompilerOutputParserCB is a callback function that is called to feed a parser
// with the plain-text compiler output.
type CompilerOutputParserCB func(cmdline []string, out []byte)

// Diagnostics represents a list of diagnostics
type Diagnostics []*Diagnostic

// Diagnostic represents a diagnostic (a compiler error, warning, note, etc.)
type Diagnostic struct {
	Severity    Severity    `json:"severity,omitempty"`
	Message     string      `json:"message"`
	File        string      `json:"file,omitempty"`
	Line        int         `json:"line,omitempty"`
	Column      int         `json:"col,omitempty"`
	Context     FullContext `json:"context,omitempty"`
	Suggestions Notes       `json:"suggestions,omitempty"`
}

// Severity is a diagnostic severity
type Severity string

const (
	// SeverityUnspecified is the undefined severity
	SeverityUnspecified Severity = ""
	// SeverityWarning is a warning
	SeverityWarning = "WARNING"
	// SeverityError is an error
	SeverityError = "ERROR"
	// SeverityFatal is a fatal error
	SeverityFatal = "FATAL"
)

// Notes represents a list of Note
type Notes []*Note

// Note represents a compiler annotation or suggestion
type Note struct {
	Message string `json:"message"`
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"col,omitempty"`
}

// FullContext represents a list of Context
type FullContext []*Context

// Context represents a context, i.e. a reference to a file, line and column
// or a part of the code that a Diagnostic refers to.
type Context struct {
	Message string `json:"message"`
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"col,omitempty"`
}

// ParseCompilerOutput parses the output of a compiler and returns a list of
// diagnostics.
func ParseCompilerOutput(compiler *DetectedCompiler, out []byte) ([]*Diagnostic, error) {
	lines := splitLines(out)
	switch compiler.Family {
	case "gcc":
		return parseGccOutput(lines)
	default:
		return nil, fmt.Errorf("unsupported compiler: %s", compiler)
	}
}

func splitLines(in []byte) []string {
	res := strings.Split(string(in), "\n")
	for i, line := range res {
		res[i] = strings.TrimSuffix(line, "\r")
	}
	if l := len(res) - 1; res[l] == "" {
		res = res[:l]
	}
	return res
}

// ToRPC converts a Diagnostics to a slice of rpc.CompileDiagnostic
func (d Diagnostics) ToRPC() []*rpc.CompileDiagnostic {
	if len(d) == 0 {
		return nil
	}
	var res []*rpc.CompileDiagnostic
	for _, diag := range d {
		res = append(res, diag.ToRPC())
	}
	return res
}

// ToRPC converts a Diagnostic to a rpc.CompileDiagnostic
func (d *Diagnostic) ToRPC() *rpc.CompileDiagnostic {
	if d == nil {
		return nil
	}
	return &rpc.CompileDiagnostic{
		Severity: string(d.Severity),
		Message:  d.Message,
		File:     d.File,
		Line:     int64(d.Line),
		Column:   int64(d.Column),
		Context:  d.Context.ToRPC(),
		Notes:    d.Suggestions.ToRPC(),
	}
}

// ToRPC converts a Notes to a slice of rpc.CompileDiagnosticNote
func (s Notes) ToRPC() []*rpc.CompileDiagnosticNote {
	var res []*rpc.CompileDiagnosticNote
	for _, suggestion := range s {
		res = append(res, suggestion.ToRPC())
	}
	return res
}

// ToRPC converts a Note to a rpc.CompileDiagnosticNote
func (s *Note) ToRPC() *rpc.CompileDiagnosticNote {
	if s == nil {
		return nil
	}
	return &rpc.CompileDiagnosticNote{
		File:    s.File,
		Line:    int64(s.Line),
		Column:  int64(s.Column),
		Message: s.Message,
	}
}

// ToRPC converts a FullContext to a slice of rpc.CompileDiagnosticContext
func (t FullContext) ToRPC() []*rpc.CompileDiagnosticContext {
	var res []*rpc.CompileDiagnosticContext
	for _, trace := range t {
		res = append(res, trace.ToRPC())
	}
	return res
}

// ToRPC converts a Context to a rpc.CompileDiagnosticContext
func (d *Context) ToRPC() *rpc.CompileDiagnosticContext {
	if d == nil {
		return nil
	}
	return &rpc.CompileDiagnosticContext{
		File:    d.File,
		Line:    int64(d.Line),
		Column:  int64(d.Column),
		Message: d.Message,
	}
}
