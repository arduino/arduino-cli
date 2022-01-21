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

package builder

import (
	"errors"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
)

type FailIfImportedLibraryIsWrong struct{}

func (s *FailIfImportedLibraryIsWrong) Run(ctx *types.Context) error {
	if len(ctx.ImportedLibraries) == 0 {
		return nil
	}

	for _, library := range ctx.ImportedLibraries {
		if !library.IsLegacy {
			if library.InstallDir.Join(constants.LIBRARY_FOLDER_ARCH).IsDir() {
				return errors.New(tr("%[1]s folder is no longer supported! See %[2]s for more information", "'arch'", "http://goo.gl/gfFJzU"))
			}
			for _, propName := range libraries.MandatoryProperties {
				if !library.Properties.ContainsKey(propName) {
					return errors.New(tr("Missing '%[1]s' from library in %[2]s", propName, library.InstallDir))
				}
			}
			if library.Layout == libraries.RecursiveLayout {
				if library.UtilityDir != nil {
					return errors.New(tr("Library can't use both '%[1]s' and '%[2]s' folders. Double check in '%[3]s'.", "src", "utility", library.InstallDir))
				}
			}
		}
	}

	return nil
}
