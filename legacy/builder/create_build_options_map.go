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
	"encoding/json"
	"strings"

	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

func CreateBuildOptionsMap(
	hardwareDirs, builtInToolsDirs, otherLibrariesDirs paths.PathList,
	builtInLibrariesDirs *paths.Path,
	sketch *sketch.Sketch,
	customBuildProperties []string,
	fqbn, compilerOptimizationFlags string,
) (string, error) {
	opts := properties.NewMap()
	opts.Set("hardwareFolders", strings.Join(hardwareDirs.AsStrings(), ","))
	opts.Set("builtInToolsFolders", strings.Join(builtInToolsDirs.AsStrings(), ","))
	if builtInLibrariesDirs != nil {
		opts.Set("builtInLibrariesFolders", builtInLibrariesDirs.String())
	}
	opts.Set("otherLibrariesFolders", strings.Join(otherLibrariesDirs.AsStrings(), ","))
	opts.SetPath("sketchLocation", sketch.FullPath)
	var additionalFilesRelative []string
	absPath := sketch.FullPath.Parent()
	for _, f := range sketch.AdditionalFiles {
		relPath, err := f.RelTo(absPath)
		if err != nil {
			continue // ignore
		}
		additionalFilesRelative = append(additionalFilesRelative, relPath.String())
	}
	opts.Set("fqbn", fqbn)
	opts.Set("customBuildProperties", strings.Join(customBuildProperties, ","))
	opts.Set("additionalFiles", strings.Join(additionalFilesRelative, ","))
	opts.Set("compiler.optimization_flags", compilerOptimizationFlags)

	buildOptionsJSON, err := json.MarshalIndent(opts, "", "  ")
	if err != nil {
		return "", errors.WithStack(err)
	}

	return string(buildOptionsJSON), nil
}
