// This file is part of arduino-cli.
//
// Copyright 2019 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.

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

	toks := strings.SplitN(arg, "@", 2)
	ret.PackageName = toks[0]
	if len(toks) > 1 {
		ret.Version = toks[1]
	}

	if parseArch {
		toks = strings.Split(ret.PackageName, ":")
		if len(toks) != 2 {
			return nil, fmt.Errorf("invalid item %s", arg)
		}
		ret.PackageName = toks[0]
		ret.Architecture = toks[1]
	}

	return ret, nil
}

// LibraryReferenceArg is a command line argument that reference a library.
type LibraryReferenceArg struct {
	Name    string
	Version string
}

func (r *LibraryReferenceArg) String() string {
	if r.Version != "" {
		return r.Name + "@" + r.Version
	}
	return r.Name
}

// ParseLibraryReferenceArg parse a command line argument that reference a
// library in the form "LibName@Version" or just "LibName".
func ParseLibraryReferenceArg(arg string) (*LibraryReferenceArg, error) {
	tokens := strings.SplitN(arg, "@", 2)

	ret := &LibraryReferenceArg{}
	// TODO: check library Name constraints
	// TODO: check library Version constraints
	if tokens[0] == "" {
		return nil, fmt.Errorf("invalid empty library name")
	}
	ret.Name = tokens[0]
	if len(tokens) > 1 {
		if tokens[1] == "" {
			return nil, fmt.Errorf("invalid empty library version")
		}
		ret.Version = tokens[1]
	}
	return ret, nil
}

// ParseLibraryReferenceArgs is a convenient wrapper that operates on a slice of strings and
// calls ParseLibraryReferenceArg for each of them. It returns at the first invalid argument.
func ParseLibraryReferenceArgs(args []string) ([]*LibraryReferenceArg, error) {
	ret := []*LibraryReferenceArg{}
	for _, arg := range args {
		if reference, err := ParseLibraryReferenceArg(arg); err == nil {
			ret = append(ret, reference)
		} else {
			return nil, err
		}
	}
	return ret, nil
}
