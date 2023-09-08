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
	"path/filepath"

	"github.com/arduino/arduino-cli/arduino/builder/utils"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

type WipeoutBuildPathIfBuildOptionsChanged struct{}

func (s *WipeoutBuildPathIfBuildOptionsChanged) Run(ctx *types.Context) error {
	if ctx.Clean {
		return doCleanup(ctx.BuildPath)
	}
	if ctx.BuildOptionsJsonPrevious == "" {
		return nil
	}
	buildOptionsJson := ctx.BuildOptionsJson
	previousBuildOptionsJson := ctx.BuildOptionsJsonPrevious

	var opts *properties.Map
	if err := json.Unmarshal([]byte(buildOptionsJson), &opts); err != nil || opts == nil {
		panic(constants.BUILD_OPTIONS_FILE + " is invalid")
	}

	var prevOpts *properties.Map
	if err := json.Unmarshal([]byte(previousBuildOptionsJson), &prevOpts); err != nil || prevOpts == nil {
		ctx.Info(tr("%[1]s invalid, rebuilding all", constants.BUILD_OPTIONS_FILE))
		return doCleanup(ctx.BuildPath)
	}

	// If SketchLocation path is different but filename is the same, consider it equal
	if filepath.Base(opts.Get("sketchLocation")) == filepath.Base(prevOpts.Get("sketchLocation")) {
		opts.Remove("sketchLocation")
		prevOpts.Remove("sketchLocation")
	}

	// If options are not changed check if core has
	if opts.Equals(prevOpts) {
		// check if any of the files contained in the core folders has changed
		// since the json was generated - like platform.txt or similar
		// if so, trigger a "safety" wipe
		buildProperties := ctx.BuildProperties
		targetCoreFolder := buildProperties.GetPath("runtime.platform.path")
		coreFolder := buildProperties.GetPath("build.core.path")
		realCoreFolder := coreFolder.Parent().Parent()
		jsonPath := ctx.BuildPath.Join(constants.BUILD_OPTIONS_FILE)
		coreUnchanged, _ := utils.DirContentIsOlderThan(realCoreFolder, jsonPath, ".txt")
		if coreUnchanged && targetCoreFolder != nil && !realCoreFolder.EqualsTo(targetCoreFolder) {
			coreUnchanged, _ = utils.DirContentIsOlderThan(targetCoreFolder, jsonPath, ".txt")
		}
		if coreUnchanged {
			return nil
		}
	}

	return doCleanup(ctx.BuildPath)
}

func doCleanup(buildPath *paths.Path) error {
	// FIXME: this should go outside legacy and behind a `logrus` call so users can
	// control when this should be printed.
	// logger.Println(constants.LOG_LEVEL_INFO, constants.MSG_BUILD_OPTIONS_CHANGED + constants.MSG_REBUILD_ALL)

	if files, err := buildPath.ReadDir(); err != nil {
		return errors.WithMessage(err, tr("cleaning build path"))
	} else {
		for _, file := range files {
			if err := file.RemoveAll(); err != nil {
				return errors.WithMessage(err, tr("cleaning build path"))
			}
		}
	}
	return nil
}
