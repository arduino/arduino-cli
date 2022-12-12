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

package lib

import (
	"context"
	"fmt"
	"strings"

	"github.com/arduino/arduino-cli/commands/lib"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

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
		return nil, fmt.Errorf(tr("invalid empty library name"))
	}
	ret.Name = tokens[0]
	if len(tokens) > 1 {
		if tokens[1] == "" {
			return nil, fmt.Errorf(tr("invalid empty library version: %s"), arg)
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

// ParseLibraryReferenceArgAndAdjustCase parse a command line argument that reference a
// library and possibly adjust the case of the name to match a library in the index
func ParseLibraryReferenceArgAndAdjustCase(instance *rpc.Instance, arg string) (*LibraryReferenceArg, error) {
	libRef, _ := ParseLibraryReferenceArg(arg)
	res, err := lib.LibrarySearch(context.Background(), &rpc.LibrarySearchRequest{
		Instance: instance,
		Query:    libRef.Name,
	})
	if err != nil {
		return nil, err
	}

	candidates := []*rpc.SearchedLibrary{}
	for _, foundLib := range res.GetLibraries() {
		if strings.EqualFold(foundLib.GetName(), libRef.Name) {
			candidates = append(candidates, foundLib)
		}
	}
	if len(candidates) == 1 {
		libRef.Name = candidates[0].GetName()
	}
	return libRef, nil
}

// ParseLibraryReferenceArgsAndAdjustCase is a convenient wrapper that operates on a slice of
// strings and calls ParseLibraryReferenceArgAndAdjustCase for each of them. It returns at the first invalid argument.
func ParseLibraryReferenceArgsAndAdjustCase(instance *rpc.Instance, args []string) ([]*LibraryReferenceArg, error) {
	ret := []*LibraryReferenceArg{}
	for _, arg := range args {
		if reference, err := ParseLibraryReferenceArgAndAdjustCase(instance, arg); err == nil {
			ret = append(ret, reference)
		} else {
			return nil, err
		}
	}
	return ret, nil
}
