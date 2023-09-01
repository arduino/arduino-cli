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

package types

import (
	"fmt"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/sketch"
	paths "github.com/arduino/go-paths-helper"
)

type SourceFile struct {
	// Path to the source file within the sketch/library root folder
	relativePath *paths.Path

	// ExtraIncludePath contains an extra include path that must be
	// used to compile this source file.
	// This is mainly used for source files that comes from old-style libraries
	// (Arduino IDE <1.5) requiring an extra include path to the "utility" folder.
	extraIncludePath *paths.Path

	// The source root for the given origin, where its source files
	// can be found. Prepending this to SourceFile.RelativePath will give
	// the full path to that source file.
	sourceRoot *paths.Path

	// The build root for the given origin, where build products will
	// be placed. Any directories inside SourceFile.RelativePath will be
	// appended here.
	buildRoot *paths.Path
}

func (f *SourceFile) Equals(g *SourceFile) bool {
	return f.relativePath.EqualsTo(g.relativePath) &&
		f.buildRoot.EqualsTo(g.buildRoot) &&
		f.sourceRoot.EqualsTo(g.sourceRoot)
}

// Create a SourceFile containing the given source file path within the
// given origin. The given path can be absolute, or relative within the
// origin's root source folder
func MakeSourceFile(ctx *Context, origin interface{}, path *paths.Path) (*SourceFile, error) {
	res := &SourceFile{}

	switch o := origin.(type) {
	case *sketch.Sketch:
		res.buildRoot = ctx.SketchBuildPath
		res.sourceRoot = ctx.SketchBuildPath
	case *libraries.Library:
		res.buildRoot = ctx.LibrariesBuildPath.Join(o.DirName)
		res.sourceRoot = o.SourceDir
		res.extraIncludePath = o.UtilityDir
	default:
		panic("Unexpected origin for SourceFile: " + fmt.Sprint(origin))
	}

	if path.IsAbs() {
		var err error
		path, err = res.sourceRoot.RelTo(path)
		if err != nil {
			return nil, err
		}
	}
	res.relativePath = path
	return res, nil
}

func (f *SourceFile) ExtraIncludePath() *paths.Path {
	return f.extraIncludePath
}

func (f *SourceFile) SourcePath() *paths.Path {
	return f.sourceRoot.JoinPath(f.relativePath)
}

func (f *SourceFile) ObjectPath() *paths.Path {
	return f.buildRoot.Join(f.relativePath.String() + ".o")
}

func (f *SourceFile) DepfilePath() *paths.Path {
	return f.buildRoot.Join(f.relativePath.String() + ".d")
}

type Command interface {
	Run(ctx *Context) error
}

type BareCommand func(ctx *Context) error

func (cmd BareCommand) Run(ctx *Context) error {
	return cmd(ctx)
}
