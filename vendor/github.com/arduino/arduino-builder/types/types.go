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

package types

import (
	"fmt"
	"strconv"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/go-paths-helper"
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
	case *Sketch:
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
	case *Sketch:
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
	Name   *paths.Path
	Source string
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

type Sketch struct {
	MainFile         SketchFile
	OtherSketchFiles []SketchFile
	AdditionalFiles  []SketchFile
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
