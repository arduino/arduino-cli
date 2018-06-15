/*
 * This file is part of Arduino Builder.
 *
 * Arduino Builder is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2015 Arduino LLC (http://www.arduino.cc/)
 */

package builder

import (
	"encoding/json"
	"path/filepath"

	"github.com/arduino/arduino-builder/builder_utils"
	"github.com/arduino/arduino-builder/constants"
	"github.com/arduino/arduino-builder/gohasissues"
	"github.com/arduino/arduino-builder/i18n"
	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/go-properties-map"
)

type WipeoutBuildPathIfBuildOptionsChanged struct{}

func (s *WipeoutBuildPathIfBuildOptionsChanged) Run(ctx *types.Context) error {
	if ctx.BuildOptionsJsonPrevious == "" {
		return nil
	}
	buildOptionsJson := ctx.BuildOptionsJson
	previousBuildOptionsJson := ctx.BuildOptionsJsonPrevious
	logger := ctx.GetLogger()

	var opts properties.Map
	var prevOpts properties.Map
	json.Unmarshal([]byte(buildOptionsJson), &opts)
	json.Unmarshal([]byte(previousBuildOptionsJson), &prevOpts)

	// If SketchLocation path is different but filename is the same, consider it equal
	if filepath.Base(opts["sketchLocation"]) == filepath.Base(prevOpts["sketchLocation"]) {
		delete(opts, "sketchLocation")
		delete(prevOpts, "sketchLocation")
	}

	// If options are not changed check if core has
	if opts.Equals(prevOpts) {
		// check if any of the files contained in the core folders has changed
		// since the json was generated - like platform.txt or similar
		// if so, trigger a "safety" wipe
		buildProperties := ctx.BuildProperties
		targetCoreFolder := buildProperties.GetPath(constants.BUILD_PROPERTIES_RUNTIME_PLATFORM_PATH)
		coreFolder := buildProperties.GetPath(constants.BUILD_PROPERTIES_BUILD_CORE_PATH)
		realCoreFolder := coreFolder.Parent().Parent()
		jsonPath := ctx.BuildPath.Join(constants.BUILD_OPTIONS_FILE)
		coreHasChanged := builder_utils.TXTBuildRulesHaveChanged(realCoreFolder, targetCoreFolder, jsonPath)

		if !coreHasChanged {
			return nil
		}
	}

	logger.Println(constants.LOG_LEVEL_INFO, constants.MSG_BUILD_OPTIONS_CHANGED)

	buildPath := ctx.BuildPath
	files, err := gohasissues.ReadDir(buildPath.String())
	if err != nil {
		return i18n.WrapError(err)
	}
	// if build path is inside the sketch folder, also wipe ctx.AdditionalFiles
	if ctx.SketchLocation != nil {
		if inside, _ := ctx.BuildPath.IsInsideDir(ctx.SketchLocation.Parent()); inside {
			ctx.Sketch.AdditionalFiles = ctx.Sketch.AdditionalFiles[:0]
		}
	}
	for _, file := range files {
		buildPath.Join(file.Name()).RemoveAll()
	}

	return nil
}
