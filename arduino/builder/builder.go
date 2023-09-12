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
	"fmt"

	"github.com/arduino/arduino-cli/arduino/builder/logger"
	"github.com/arduino/arduino-cli/arduino/builder/progress"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
)

// ErrSketchCannotBeLocatedInBuildPath fixdoc
var ErrSketchCannotBeLocatedInBuildPath = errors.New("sketch cannot be located in build path")

// Builder is a Sketch builder.
type Builder struct {
	sketch          *sketch.Sketch
	buildProperties *properties.Map

	buildPath          *paths.Path
	sketchBuildPath    *paths.Path
	coreBuildPath      *paths.Path
	librariesBuildPath *paths.Path

	// Parallel processes
	jobs int

	// Custom build properties defined by user (line by line as "key=value" pairs)
	customBuildProperties []string

	// core related
	coreBuildCachePath *paths.Path

	logger *logger.BuilderLogger
	clean  bool

	// Source code overrides (filename -> content map).
	// The provided source data is used instead of reading it from disk.
	// The keys of the map are paths relative to sketch folder.
	sourceOverrides map[string]string

	// Set to true to skip build and produce only Compilation Database
	onlyUpdateCompilationDatabase bool

	// Progress of all various steps
	Progress *progress.Struct

	*BuildOptionsManager
}

// NewBuilder creates a sketch Builder.
func NewBuilder(
	sk *sketch.Sketch,
	boardBuildProperties *properties.Map,
	buildPath *paths.Path,
	optimizeForDebug bool,
	coreBuildCachePath *paths.Path,
	jobs int,
	requestBuildProperties []string,
	hardwareDirs, builtInToolsDirs, otherLibrariesDirs paths.PathList,
	builtInLibrariesDirs *paths.Path,
	fqbn *cores.FQBN,
	clean bool,
	sourceOverrides map[string]string,
	onlyUpdateCompilationDatabase bool,
	logger *logger.BuilderLogger,
	progressStats *progress.Struct,
) (*Builder, error) {
	buildProperties := properties.NewMap()
	if boardBuildProperties != nil {
		buildProperties.Merge(boardBuildProperties)
	}

	if buildPath != nil {
		buildProperties.SetPath("build.path", buildPath)
	}
	if sk != nil {
		buildProperties.Set("build.project_name", sk.MainFile.Base())
		buildProperties.SetPath("build.source.path", sk.FullPath)
	}
	if optimizeForDebug {
		if debugFlags, ok := buildProperties.GetOk("compiler.optimization_flags.debug"); ok {
			buildProperties.Set("compiler.optimization_flags", debugFlags)
		}
	} else {
		if releaseFlags, ok := buildProperties.GetOk("compiler.optimization_flags.release"); ok {
			buildProperties.Set("compiler.optimization_flags", releaseFlags)
		}
	}

	// Add user provided custom build properties
	customBuildProperties, err := properties.LoadFromSlice(requestBuildProperties)
	if err != nil {
		return nil, fmt.Errorf("invalid build properties: %w", err)
	}
	buildProperties.Merge(customBuildProperties)
	customBuildPropertiesArgs := append(requestBuildProperties, "build.warn_data_percentage=75")

	sketchBuildPath, err := buildPath.Join("sketch").Abs()
	if err != nil {
		return nil, err
	}
	librariesBuildPath, err := buildPath.Join("libraries").Abs()
	if err != nil {
		return nil, err
	}
	coreBuildPath, err := buildPath.Join("core").Abs()
	if err != nil {
		return nil, err
	}

	if buildPath.Canonical().EqualsTo(sk.FullPath.Canonical()) {
		return nil, ErrSketchCannotBeLocatedInBuildPath
	}

	if progressStats == nil {
		progressStats = progress.New(nil)
	}

	return &Builder{
		sketch:                        sk,
		buildProperties:               buildProperties,
		buildPath:                     buildPath,
		sketchBuildPath:               sketchBuildPath,
		coreBuildPath:                 coreBuildPath,
		librariesBuildPath:            librariesBuildPath,
		jobs:                          jobs,
		customBuildProperties:         customBuildPropertiesArgs,
		coreBuildCachePath:            coreBuildCachePath,
		logger:                        logger,
		clean:                         clean,
		sourceOverrides:               sourceOverrides,
		onlyUpdateCompilationDatabase: onlyUpdateCompilationDatabase,
		Progress:                      progressStats,
		BuildOptionsManager: NewBuildOptionsManager(
			hardwareDirs, builtInToolsDirs, otherLibrariesDirs,
			builtInLibrariesDirs, buildPath,
			sk,
			customBuildPropertiesArgs,
			fqbn,
			clean,
			buildProperties.Get("compiler.optimization_flags"),
			buildProperties.GetPath("runtime.platform.path"),
			buildProperties.GetPath("build.core.path"), // TODO can we buildCorePath ?
			logger,
		),
	}, nil
}

// GetBuildProperties returns the build properties for running this build
func (b *Builder) GetBuildProperties() *properties.Map {
	return b.buildProperties
}

// GetBuildPath returns the build path
func (b *Builder) GetBuildPath() *paths.Path {
	return b.buildPath
}

// GetSketchBuildPath returns the sketch build path
func (b *Builder) GetSketchBuildPath() *paths.Path {
	return b.sketchBuildPath
}

// GetLibrariesBuildPath returns the libraries build path
func (b *Builder) GetLibrariesBuildPath() *paths.Path {
	return b.librariesBuildPath
}
