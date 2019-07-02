// This file is part of arduino-cli.
//
// Copyright 2019 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.

package types

import (
	"fmt"

	"github.com/arduino/arduino-cli/arduino/libraries"
	paths "github.com/arduino/go-paths-helper"
)

// SourceFile represents a source file
type SourceFile struct {
	// Sketch or Library pointer that this source file lives in
	Origin interface{}
	// Path to the source file within the sketch/library root folder
	RelativePath *paths.Path
}

// NewSourceFile creates a SourceFile containing the given source file path within the
// given origin. The given path can be absolute, or relative within the
// origin's root source folder
func NewSourceFile(ctx *Context, origin interface{}, path *paths.Path) (SourceFile, error) {
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

// SourcePath returns the full path to the source file
func (f *SourceFile) SourcePath(ctx *Context) *paths.Path {
	return sourceRoot(ctx, f.Origin).JoinPath(f.RelativePath)
}

// ObjectPath returns the full path to the .o file for this source
func (f *SourceFile) ObjectPath(ctx *Context) *paths.Path {
	return buildRoot(ctx, f.Origin).Join(f.RelativePath.String() + ".o")
}

// DepfilePath returns the full path to the .d file for this source
func (f *SourceFile) DepfilePath(ctx *Context) *paths.Path {
	return buildRoot(ctx, f.Origin).Join(f.RelativePath.String() + ".d")
}

// UniqueSourceFileQueue is a slice of SourceFile
type UniqueSourceFileQueue []SourceFile

func (queue UniqueSourceFileQueue) Len() int           { return len(queue) }
func (queue UniqueSourceFileQueue) Less(i, j int) bool { return false }
func (queue UniqueSourceFileQueue) Swap(i, j int)      { panic("Who called me?!?") }

// Push add an item if not already present
func (queue *UniqueSourceFileQueue) Push(value SourceFile) {
	found := false
	for _, elem := range *queue {
		if elem.Origin == value.Origin && elem.RelativePath.EqualsTo(value.RelativePath) {
			found = true
			break
		}
	}

	if !found {
		*queue = append(*queue, value)
	}
}

// Pop returns the first item of the slice and removes it
func (queue *UniqueSourceFileQueue) Pop() SourceFile {
	old := *queue
	x := old[0]
	*queue = old[1:]
	return x
}

// Empty returns whether the slice is empty
func (queue *UniqueSourceFileQueue) Empty() bool {
	return queue.Len() == 0
}
