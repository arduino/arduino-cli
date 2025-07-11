// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
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

package detector

import (
	"slices"

	"github.com/arduino/go-paths-helper"
)

type sourceFile struct {
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

// Equals fixdoc
func (f *sourceFile) Equals(g *sourceFile) bool {
	return f.relativePath.EqualsTo(g.relativePath) &&
		f.buildRoot.EqualsTo(g.buildRoot) &&
		f.sourceRoot.EqualsTo(g.sourceRoot)
}

// makeSourceFile create a sourceFile object for the given source file path.
// The given sourceFilePath can be absolute, or relative within the sourceRoot root folder.
func makeSourceFile(sourceRoot, buildRoot, sourceFilePath *paths.Path, extraIncludePath ...*paths.Path) (*sourceFile, error) {
	res := &sourceFile{
		buildRoot:  buildRoot,
		sourceRoot: sourceRoot,
	}

	if len(extraIncludePath) > 1 {
		panic("only one extra include path allowed")
	}
	if len(extraIncludePath) > 0 {
		res.extraIncludePath = extraIncludePath[0]
	}

	if sourceFilePath.IsAbs() {
		var err error
		sourceFilePath, err = res.sourceRoot.RelTo(sourceFilePath)
		if err != nil {
			return nil, err
		}
	}
	res.relativePath = sourceFilePath
	return res, nil
}

// ExtraIncludePath returns the extra include path required to build the source file.
func (f *sourceFile) ExtraIncludePath() *paths.Path {
	return f.extraIncludePath
}

// SourcePath return the full path to the source file.
func (f *sourceFile) SourcePath() *paths.Path {
	return f.sourceRoot.JoinPath(f.relativePath)
}

// ObjectPath return the full path to the object file.
func (f *sourceFile) ObjectPath() *paths.Path {
	return f.buildRoot.Join(f.relativePath.String() + ".o")
}

// DepfilePath return the full path to the dependency file.
func (f *sourceFile) DepfilePath() *paths.Path {
	return f.buildRoot.Join(f.relativePath.String() + ".d")
}

// uniqueSourceFileQueue is a queue of source files that does not allow duplicates.
type uniqueSourceFileQueue []*sourceFile

// Push adds a source file to the queue if it is not already present.
func (queue *uniqueSourceFileQueue) Push(value *sourceFile) {
	if !queue.Contains(value) {
		*queue = append(*queue, value)
	}
}

// Contains checks if the queue Contains a source file.
func (queue uniqueSourceFileQueue) Contains(target *sourceFile) bool {
	return slices.ContainsFunc(queue, target.Equals)
}

// Pop removes and returns the first element of the queue.
func (queue *uniqueSourceFileQueue) Pop() *sourceFile {
	old := *queue
	x := old[0]
	*queue = old[1:]
	return x
}

// Empty returns true if the queue is Empty.
func (queue uniqueSourceFileQueue) Empty() bool {
	return len(queue) == 0
}
