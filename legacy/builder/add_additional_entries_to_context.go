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
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
)

func AddAdditionalEntriesToContext(buildPath *paths.Path, warningLevel string) (*paths.Path, *paths.Path, *paths.Path, string, error) {
	sketchBuildPath, err := buildPath.Join(constants.FOLDER_SKETCH).Abs()
	if err != nil {
		return nil, nil, nil, "", errors.WithStack(err)
	}
	librariesBuildPath, err := buildPath.Join(constants.FOLDER_LIBRARIES).Abs()
	if err != nil {
		return nil, nil, nil, "", errors.WithStack(err)
	}
	coreBuildPath, err := buildPath.Join(constants.FOLDER_CORE).Abs()
	if err != nil {
		return nil, nil, nil, "", errors.WithStack(err)
	}

	if warningLevel == "" {
		warningLevel = DEFAULT_WARNINGS_LEVEL
	}

	return sketchBuildPath, librariesBuildPath, coreBuildPath, warningLevel, nil
}
