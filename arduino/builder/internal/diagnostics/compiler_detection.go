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
	"bytes"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-cli/executils"
	semver "go.bug.st/relaxed-semver"
)

// DetectedCompiler represents a compiler detected from a given command line
type DetectedCompiler struct {
	Name            string
	Family          string
	Version         *semver.Version
	DetailedVersion []string
}

// This function is overridden for mocking unit tests
var runProcess = func(args ...string) []string {
	if cmd, err := executils.NewProcess(nil, args...); err == nil {
		out := &bytes.Buffer{}
		cmd.RedirectStdoutTo(out)
		cmd.Run()
		return splitLines(out.Bytes())
	}
	return nil
}

// DetectCompilerFromCommandLine tries to detect a compiler from a given command line.
// If probeCompiler is true, the compiler may be executed with different flags to
// infer the version or capabilities.
func DetectCompilerFromCommandLine(args []string, probeCompiler bool) *DetectedCompiler {
	if len(args) == 0 {
		return nil
	}
	basename := filepath.Base(args[0])
	family := ""
	if strings.Contains(basename, "g++") || strings.Contains(basename, "gcc") {
		family = "gcc"
	}
	res := &DetectedCompiler{
		Name:   basename,
		Family: family,
	}

	if family == "gcc" && probeCompiler {
		// Run "gcc --version" to obtain more info
		res.DetailedVersion = runProcess(args[0], "--version")

		// Usually on the first line we get the compiler architecture and
		// version (as last field), followed by the compiler license, for
		// example:
		//
		//   g++ (Ubuntu 12.2.0-3ubuntu1) 12.2.0
		//   Copyright (C) 2022 Free Software Foundation, Inc.
		//   This is free software; see the source for copying conditions.  There is NO
		//   warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
		//
		if len(res.DetailedVersion) > 0 {
			split := strings.Split(res.DetailedVersion[0], " ")
			if len(split) >= 3 {
				res.Name = split[0]
				res.Version, _ = semver.Parse(split[len(split)-1])
			}
		}
	}
	return res
}
