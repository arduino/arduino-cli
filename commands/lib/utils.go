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
	"fmt"

	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/commands"
)

type libraryReferencer interface {
	commands.Versioned
	GetName() string
}

func createLibIndexReference(lm *librariesmanager.LibrariesManager, req libraryReferencer) (*librariesindex.Reference, error) {
	version, err := commands.ParseVersion(req)
	if err != nil {
		return nil, fmt.Errorf("invalid version: %s", err)
	}

	return &librariesindex.Reference{Name: req.GetName(), Version: version}, nil
}

func findLibraryIndexRelease(lm *librariesmanager.LibrariesManager, req libraryReferencer) (*librariesindex.Release, error) {
	ref, err := createLibIndexReference(lm, req)
	if err != nil {
		return nil, err
	}
	lib := lm.Index.FindRelease(ref)
	if lib == nil {
		return nil, fmt.Errorf("library %s not found", ref)
	}
	return lib, nil
}
