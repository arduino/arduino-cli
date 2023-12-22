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
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesmanager"
)

type libraryReferencer interface {
	commands.Versioned
	GetName() string
}

func createLibIndexReference(req libraryReferencer) (*librariesindex.Reference, error) {
	version, err := commands.ParseVersion(req)
	if err != nil {
		return nil, &cmderrors.InvalidVersionError{Cause: err}
	}

	return &librariesindex.Reference{Name: req.GetName(), Version: version}, nil
}

func findLibraryIndexRelease(lm *librariesmanager.LibrariesManager, req libraryReferencer) (*librariesindex.Release, error) {
	ref, err := createLibIndexReference(req)
	if err != nil {
		return nil, err
	}
	lib := lm.Index.FindRelease(ref)
	if lib == nil {
		return nil, &cmderrors.LibraryNotFoundError{Library: ref.String()}
	}
	return lib, nil
}
