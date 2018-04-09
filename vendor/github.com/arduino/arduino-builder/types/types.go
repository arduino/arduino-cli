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
	"path/filepath"
	"strconv"

	"github.com/arduino/arduino-builder/constants"
)

type SourceFile struct {
	// Sketch or Library pointer that this source file lives in
	Origin interface{}
	// Path to the source file within the sketch/library root folder
	RelativePath string
}

// Create a SourceFile containing the given source file path within the
// given origin. The given path can be absolute, or relative within the
// origin's root source folder
func MakeSourceFile(ctx *Context, origin interface{}, path string) (SourceFile, error) {
	if filepath.IsAbs(path) {
		var err error
		path, err = filepath.Rel(sourceRoot(ctx, origin), path)
		if err != nil {
			return SourceFile{}, err
		}
	}
	return SourceFile{Origin: origin, RelativePath: path}, nil
}

// Return the build root for the given origin, where build products will
// be placed. Any directories inside SourceFile.RelativePath will be
// appended here.
func buildRoot(ctx *Context, origin interface{}) string {
	switch o := origin.(type) {
	case *Sketch:
		return ctx.SketchBuildPath
	case *Library:
		return filepath.Join(ctx.LibrariesBuildPath, o.Name)
	default:
		panic("Unexpected origin for SourceFile: " + fmt.Sprint(origin))
	}
}

// Return the source root for the given origin, where its source files
// can be found. Prepending this to SourceFile.RelativePath will give
// the full path to that source file.
func sourceRoot(ctx *Context, origin interface{}) string {
	switch o := origin.(type) {
	case *Sketch:
		return ctx.SketchBuildPath
	case *Library:
		return o.SrcFolder
	default:
		panic("Unexpected origin for SourceFile: " + fmt.Sprint(origin))
	}
}

func (f *SourceFile) SourcePath(ctx *Context) string {
	return filepath.Join(sourceRoot(ctx, f.Origin), f.RelativePath)
}

func (f *SourceFile) ObjectPath(ctx *Context) string {
	return filepath.Join(buildRoot(ctx, f.Origin), f.RelativePath+".o")
}

func (f *SourceFile) DepfilePath(ctx *Context) string {
	return filepath.Join(buildRoot(ctx, f.Origin), f.RelativePath+".d")
}

type SketchFile struct {
	Name   string
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
	return s[i].Name < s[j].Name
}

type Sketch struct {
	MainFile         SketchFile
	OtherSketchFiles []SketchFile
	AdditionalFiles  []SketchFile
}

type LibraryLayout uint16

const (
	LIBRARY_FLAT LibraryLayout = 1 << iota
	LIBRARY_RECURSIVE
)

type Library struct {
	Folder        string
	SrcFolder     string
	UtilityFolder string
	Layout        LibraryLayout
	Name          string
	RealName      string
	Archs         []string
	DotALinkage   bool
	Precompiled   bool
	LDflags       string
	IsLegacy      bool
	Version       string
	Author        string
	Maintainer    string
	Sentence      string
	Paragraph     string
	URL           string
	Category      string
	License       string
	Properties    map[string]string
}

func (library *Library) String() string {
	return library.Name + " : " + library.SrcFolder
}

func (library *Library) SupportsArchitectures(archs []string) bool {
	if sliceContains(archs, constants.LIBRARY_ALL_ARCHS) || sliceContains(library.Archs, constants.LIBRARY_ALL_ARCHS) {
		return true
	}

	for _, libraryArch := range library.Archs {
		if sliceContains(archs, libraryArch) {
			return true
		}
	}

	return false
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

type SourceFolder struct {
	Folder  string
	Recurse bool
}

type LibraryResolutionResult struct {
	Library          *Library
	NotUsedLibraries []*Library
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

func LibraryToSourceFolder(library *Library) []SourceFolder {
	sourceFolders := []SourceFolder{}
	recurse := library.Layout == LIBRARY_RECURSIVE
	sourceFolders = append(sourceFolders, SourceFolder{Folder: library.SrcFolder, Recurse: recurse})
	if library.UtilityFolder != "" {
		sourceFolders = append(sourceFolders, SourceFolder{Folder: library.UtilityFolder, Recurse: false})
	}
	return sourceFolders
}

type Command interface {
	Run(ctx *Context) error
}
