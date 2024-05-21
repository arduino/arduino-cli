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
	"fmt"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/utils"
	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/internal/arduino/sketch"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
)

// buildOptions fixdoc
type buildOptions struct {
	currentOptions *properties.Map

	hardwareDirs              paths.PathList
	otherLibrariesDirs        paths.PathList
	builtInLibrariesDirs      *paths.Path
	buildPath                 *paths.Path
	runtimePlatformPath       *paths.Path
	buildCorePath             *paths.Path
	sketch                    *sketch.Sketch
	customBuildProperties     []string
	compilerOptimizationFlags string
	clean                     bool
}

// newBuildOptions fixdoc
func newBuildOptions(
	hardwareDirs, otherLibrariesDirs paths.PathList,
	builtInLibrariesDirs, buildPath *paths.Path,
	sketch *sketch.Sketch,
	customBuildProperties []string,
	fqbn *cores.FQBN,
	clean bool,
	compilerOptimizationFlags string,
	runtimePlatformPath, buildCorePath *paths.Path,
) *buildOptions {
	opts := properties.NewMap()

	opts.Set("hardwareFolders", strings.Join(hardwareDirs.AsStrings(), ","))
	opts.Set("otherLibrariesFolders", strings.Join(otherLibrariesDirs.AsStrings(), ","))
	opts.SetPath("sketchLocation", sketch.FullPath)
	opts.Set("fqbn", fqbn.String())
	opts.Set("customBuildProperties", strings.Join(customBuildProperties, ","))
	opts.Set("compiler.optimization_flags", compilerOptimizationFlags)

	if builtInLibrariesDirs != nil {
		opts.Set("builtInLibrariesFolders", builtInLibrariesDirs.String())
	}

	absPath := sketch.FullPath.Parent()
	var additionalFilesRelative []string
	for _, f := range sketch.AdditionalFiles {
		relPath, err := f.RelTo(absPath)
		if err != nil {
			continue // ignore
		}
		additionalFilesRelative = append(additionalFilesRelative, relPath.String())
	}
	opts.Set("additionalFiles", strings.Join(additionalFilesRelative, ","))

	return &buildOptions{
		currentOptions:            opts,
		hardwareDirs:              hardwareDirs,
		otherLibrariesDirs:        otherLibrariesDirs,
		builtInLibrariesDirs:      builtInLibrariesDirs,
		buildPath:                 buildPath,
		runtimePlatformPath:       runtimePlatformPath,
		buildCorePath:             buildCorePath,
		sketch:                    sketch,
		customBuildProperties:     customBuildProperties,
		compilerOptimizationFlags: compilerOptimizationFlags,
		clean:                     clean,
	}
}

func (b *Builder) createBuildOptionsJSON() error {
	buildOptionsJSON, err := json.MarshalIndent(b.buildOptions.currentOptions, "", "  ")
	if err != nil {
		return err
	}
	return b.buildOptions.buildPath.Join("build.options.json").WriteFile(buildOptionsJSON)
}

func (b *Builder) wipeBuildPath() error {
	// FIXME: this should go outside legacy and behind a `logrus` call so users can
	// control when this should be printed.
	// logger.Println(constants.LOG_LEVEL_INFO, constants.MSG_BUILD_OPTIONS_CHANGED + constants.MSG_REBUILD_ALL)
	if err := b.buildOptions.buildPath.RemoveAll(); err != nil {
		return fmt.Errorf("%s: %w", i18n.Tr("cleaning build path"), err)
	}
	if err := b.buildOptions.buildPath.MkdirAll(); err != nil {
		return fmt.Errorf("%s: %w", i18n.Tr("cleaning build path"), err)
	}
	return nil
}

func (b *Builder) wipeBuildPathIfBuildOptionsChanged() error {
	if b.buildOptions.clean {
		return b.wipeBuildPath()
	}

	// Load previous build options map
	var buildOptionsJSONPrevious []byte
	if buildOptionsFile := b.buildOptions.buildPath.Join("build.options.json"); buildOptionsFile.Exist() {
		var err error
		buildOptionsJSONPrevious, err = buildOptionsFile.ReadFile()
		if err != nil {
			return err
		}
	}

	if len(buildOptionsJSONPrevious) == 0 {
		return nil
	}

	var prevOpts *properties.Map
	if err := json.Unmarshal(buildOptionsJSONPrevious, &prevOpts); err != nil || prevOpts == nil {
		b.logger.Info(i18n.Tr("%[1]s invalid, rebuilding all", "build.options.json"))
		return b.wipeBuildPath()
	}

	// Since we might apply a side effect we clone it
	currentOptions := b.buildOptions.currentOptions.Clone()
	// If SketchLocation path is different but filename is the same, consider it equal
	if filepath.Base(currentOptions.Get("sketchLocation")) == filepath.Base(prevOpts.Get("sketchLocation")) {
		currentOptions.Remove("sketchLocation")
		prevOpts.Remove("sketchLocation")
	}

	// If options are not changed check if core has
	if currentOptions.Equals(prevOpts) {
		// check if any of the files contained in the core folders has changed
		// since the json was generated - like platform.txt or similar
		// if so, trigger a "safety" wipe
		targetCoreFolder := b.buildOptions.runtimePlatformPath
		coreFolder := b.buildOptions.buildCorePath
		realCoreFolder := coreFolder.Parent().Parent()
		jsonPath := b.buildOptions.buildPath.Join("build.options.json")
		coreUnchanged, _ := utils.DirContentIsOlderThan(realCoreFolder, jsonPath, ".txt")
		if coreUnchanged && targetCoreFolder != nil && !realCoreFolder.EqualsTo(targetCoreFolder) {
			coreUnchanged, _ = utils.DirContentIsOlderThan(targetCoreFolder, jsonPath, ".txt")
		}
		if coreUnchanged {
			return nil
		}
	}

	return b.wipeBuildPath()
}
