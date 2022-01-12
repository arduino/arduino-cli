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
	"strconv"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/sketch"
	paths "github.com/arduino/go-paths-helper"
)

type SourceFile struct {
	// Sketch or Library pointer that this source file lives in
	Origin interface{}
	// Path to the source file within the sketch/library root folder
	RelativePath *paths.Path
}

// Create a SourceFile containing the given source file path within the
// given origin. The given path can be absolute, or relative within the
// origin's root source folder
func MakeSourceFile(ctx *Context, origin interface{}, path *paths.Path) (SourceFile, error) {
	if path.IsAbs() {
		var err error
		path, err = sourceRoot(ctx, origin).RelTo(path)
		if err != nil {
			return SourceFile{}, err
		}
	}
	return SourceFile{Origin: origin, RelativePath: path}, nil
}

// Return the build root for the given origin, where build products will
// be placed. Any directories inside SourceFile.RelativePath will be
// appended here.
func buildRoot(ctx *Context, origin interface{}) *paths.Path {
	switch o := origin.(type) {
	case *sketch.Sketch:
		return ctx.SketchBuildPath
	case *libraries.Library:
		return ctx.LibrariesBuildPath.Join(o.Name)
	default:
		panic("Unexpected origin for SourceFile: " + fmt.Sprint(origin))
	}
}

// Return the source root for the given origin, where its source files
// can be found. Prepending this to SourceFile.RelativePath will give
// the full path to that source file.
func sourceRoot(ctx *Context, origin interface{}) *paths.Path {
	switch o := origin.(type) {
	case *sketch.Sketch:
		return ctx.SketchBuildPath
	case *libraries.Library:
		return o.SourceDir
	default:
		panic("Unexpected origin for SourceFile: " + fmt.Sprint(origin))
	}
}

func (f *SourceFile) SourcePath(ctx *Context) *paths.Path {
	return sourceRoot(ctx, f.Origin).JoinPath(f.RelativePath)
}

func (f *SourceFile) ObjectPath(ctx *Context) *paths.Path {
	return buildRoot(ctx, f.Origin).Join(f.RelativePath.String() + ".o")
}

func (f *SourceFile) DepfilePath(ctx *Context) *paths.Path {
	return buildRoot(ctx, f.Origin).Join(f.RelativePath.String() + ".d")
}

type SketchFile struct {
	Name *paths.Path
}

type SketchFileSortByName []SketchFile

func (s SketchFileSortByName) Len() int {
	return len(s)
}

func (s SketchFileSortByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SketchFileSortByName) Less(i, j int) bool {
	return s[i].Name.String() < s[j].Name.String()
}

type PlatforKeysRewrite struct {
	Rewrites []PlatforKeyRewrite
}

func (p *PlatforKeysRewrite) Empty() bool {
	return len(p.Rewrites) == 0
}

type PlatforKeyRewrite struct {
	Key      string
	OldValue string
	NewValue string
}

type Prototype struct {
	FunctionName string
	File         string
	Prototype    string
	Modifiers    string
	Line         int
}

func (proto *Prototype) String() string {
	return proto.Modifiers + " " + proto.Prototype + " @ " + strconv.Itoa(proto.Line)
}

type LibraryResolutionResult struct {
	Library          *libraries.Library
	NotUsedLibraries []*libraries.Library
}

type CTag struct {
	FunctionName string
	Kind         string
	Line         int
	Code         string
	Class        string
	Struct       string
	Namespace    string
	Filename     string
	Typeref      string
	SkipMe       bool
	Signature    string

	Prototype          string
	PrototypeModifiers string
}

type Command interface {
	Run(ctx *Context) error
}
