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

package globals

import (
	"fmt"
	"strings"
)

// ReferenceArg represents a reference item (core or library) passed to the CLI
// interface
type ReferenceArg struct {
	PackageName  string
	Architecture string
	Version      string
}

func (r *ReferenceArg) String() string {
	if r.Version != "" {
		return r.PackageName + ":" + r.Architecture + "@" + r.Version
	}
	return r.PackageName + ":" + r.Architecture
}

// ParseReferenceArgs is a convenient wrapper that operates on a slice of strings and
// calls ParseReferenceArg for each of them. It returns at the first invalid argument.
func ParseReferenceArgs(args []string, parseArch bool) ([]*ReferenceArg, error) {
	ret := []*ReferenceArg{}
	for _, arg := range args {
		reference, err := ParseReferenceArg(arg, parseArch)
		if err != nil {
			return nil, err
		}
		ret = append(ret, reference)
	}
	return ret, nil
}

// ParseReferenceArg parses a string and return a ReferenceArg object. If `parseArch` is passed,
// the method also tries to parse the architecture bit, i.e. string must be in the form
// "packager:arch@version", useful to represent a platform (or core) name.
func ParseReferenceArg(arg string, parseArch bool) (*ReferenceArg, error) {
	ret := &ReferenceArg{}
	if arg == "" {
		return nil, fmt.Errorf("invalid empty core argument")
	}
	toks := strings.SplitN(arg, "@", 2)
	if toks[0] == "" {
		return nil, fmt.Errorf("invalid empty core reference '%s'", arg)
	}
	ret.PackageName = toks[0]
	if len(toks) > 1 {
		if toks[1] == "" {
			return nil, fmt.Errorf("invalid empty core version: '%s'", arg)
		}
		ret.Version = toks[1]
	}

	if parseArch {
		toks = strings.Split(ret.PackageName, ":")
		if len(toks) != 2 {
			return nil, fmt.Errorf("invalid item %s", arg)
		}
		if toks[0] == "" {
			return nil, fmt.Errorf("invalid empty core name '%s'", arg)
		}
		ret.PackageName = toks[0]
		if toks[1] == "" {
			return nil, fmt.Errorf("invalid empty core architecture '%s'", arg)
		}
		ret.Architecture = toks[1]
	}

	return ret, nil
}
