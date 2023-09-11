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
	"github.com/arduino/go-properties-orderedmap"
)

// Builder is a Sketch builder.
type Builder struct {
	sketch          *sketch.Sketch
	buildProperties *properties.Map

	// Parallel processes
	jobs int

	// core related
	coreBuildCachePath *paths.Path
}

// NewBuilder creates a sketch Builder.
func NewBuilder(
	sk *sketch.Sketch,
	boardBuildProperties *properties.Map,
	buildPath *paths.Path,
	optimizeForDebug bool,
	coreBuildCachePath *paths.Path,
	jobs int,
) *Builder {
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

	return &Builder{
		sketch:             sk,
		buildProperties:    buildProperties,
		coreBuildCachePath: coreBuildCachePath,
		jobs:               jobs,
	}
}

// GetBuildProperties returns the build properties for running this build
func (b *Builder) GetBuildProperties() *properties.Map {
	return b.buildProperties
}

// Jobs number of parallel processes
func (b *Builder) Jobs() int {
	return b.jobs
}
