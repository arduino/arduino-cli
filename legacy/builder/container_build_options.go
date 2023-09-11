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
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

func ContainerBuildOptions(
	hardwareDirs, builtInToolsDirs, otherLibrariesDirs paths.PathList,
	builtInLibrariesDirs, buildPath *paths.Path,
	sketch *sketch.Sketch,
	customBuildProperties []string,
	fqbn string,
	clean bool,
	buildProperties *properties.Map,
) (string, string, string, error) {
	buildOptionsJSON, err := CreateBuildOptionsMap(
		hardwareDirs, builtInToolsDirs, otherLibrariesDirs,
		builtInLibrariesDirs, sketch, customBuildProperties,
		fqbn, buildProperties.Get("compiler.optimization_flags"),
	)
	if err != nil {
		return "", "", "", errors.WithStack(err)
	}

	buildOptionsJSONPrevious, err := LoadPreviousBuildOptionsMap(buildPath)
	if err != nil {
		return "", "", "", errors.WithStack(err)
	}

	infoOut, err := WipeoutBuildPathIfBuildOptionsChanged(
		clean,
		buildPath,
		buildOptionsJSON,
		buildOptionsJSONPrevious,
		buildProperties,
	)
	if err != nil {
		return "", "", "", errors.WithStack(err)
	}

	return buildOptionsJSON, buildOptionsJSONPrevious, infoOut, StoreBuildOptionsMap(buildPath, buildOptionsJSON)
}
