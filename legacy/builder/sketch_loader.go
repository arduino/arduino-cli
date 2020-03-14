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
	"sort"
	"strings"

	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/i18n"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	"github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
)

type SketchLoader struct{}

func (s *SketchLoader) Run(ctx *types.Context) error {
	if ctx.SketchLocation == nil {
		return nil
	}

	sketchLocation := ctx.SketchLocation

	sketchLocation, err := sketchLocation.Abs()
	if err != nil {
		return errors.WithStack(err)
	}
	mainSketchStat, err := sketchLocation.Stat()
	if err != nil {
		return errors.WithStack(err)
	}
	if mainSketchStat.IsDir() {
		sketchLocation = sketchLocation.Join(mainSketchStat.Name() + ".ino")
	}

	ctx.SketchLocation = sketchLocation

	allSketchFilePaths, err := collectAllSketchFiles(sketchLocation.Parent())
	if err != nil {
		return errors.WithStack(err)
	}

	logger := ctx.GetLogger()

	if !allSketchFilePaths.Contains(sketchLocation) {
		return i18n.ErrorfWithLogger(logger, constants.MSG_CANT_FIND_SKETCH_IN_PATH, sketchLocation, sketchLocation.Parent())
	}

	sketch, err := makeSketch(sketchLocation, allSketchFilePaths, ctx.BuildPath, logger)
	if err != nil {
		return errors.WithStack(err)
	}

	ctx.SketchLocation = sketchLocation
	ctx.Sketch = sketch

	return nil
}

func collectAllSketchFiles(from *paths.Path) (paths.PathList, error) {
	filePaths := []string{}
	// Source files in the root are compiled, non-recursively. This
	// is the only place where .ino files can be present.
	rootExtensions := func(ext string) bool { return MAIN_FILE_VALID_EXTENSIONS[ext] || ADDITIONAL_FILE_VALID_EXTENSIONS[ext] }
	err := utils.FindFilesInFolder(&filePaths, from.String(), rootExtensions, true /* recurse */)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return paths.NewPathList(filePaths...), errors.WithStack(err)
}

func makeSketch(sketchLocation *paths.Path, allSketchFilePaths paths.PathList, buildLocation *paths.Path, logger i18n.Logger) (*types.Sketch, error) {
	sketchFilesMap := make(map[string]types.SketchFile)
	for _, sketchFilePath := range allSketchFilePaths {
		sketchFilesMap[sketchFilePath.String()] = types.SketchFile{Name: sketchFilePath}
	}

	mainFile := sketchFilesMap[sketchLocation.String()]
	delete(sketchFilesMap, sketchLocation.String())

	additionalFiles := []types.SketchFile{}
	otherSketchFiles := []types.SketchFile{}
	mainFileDir := mainFile.Name.Parent()
	for _, sketchFile := range sketchFilesMap {
		ext := strings.ToLower(sketchFile.Name.Ext())
		if MAIN_FILE_VALID_EXTENSIONS[ext] {
			if sketchFile.Name.Parent().EqualsTo(mainFileDir) {
				otherSketchFiles = append(otherSketchFiles, sketchFile)
			}
		} else if ADDITIONAL_FILE_VALID_EXTENSIONS[ext] {
			if buildLocation == nil || !strings.Contains(sketchFile.Name.Parent().String(), buildLocation.String()) {
				additionalFiles = append(additionalFiles, sketchFile)
			}
		} else {
			return nil, i18n.ErrorfWithLogger(logger, constants.MSG_UNKNOWN_SKETCH_EXT, sketchFile.Name)
		}
	}

	sort.Sort(types.SketchFileSortByName(additionalFiles))
	sort.Sort(types.SketchFileSortByName(otherSketchFiles))

	return &types.Sketch{MainFile: mainFile, OtherSketchFiles: otherSketchFiles, AdditionalFiles: additionalFiles}, nil
}
