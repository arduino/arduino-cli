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
	"fmt"
	"slices"

	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/utils"
	"github.com/arduino/go-paths-helper"
)

type sourceFile struct {
	// SourcePath is the path to the source file
	SourcePath *paths.Path `json:"source_path"`

	// ObjectPath is the path to the object file that will be generated
	ObjectPath *paths.Path `json:"object_path"`

	// DepfilePath is the path to the dependency file that will be generated
	DepfilePath *paths.Path `json:"depfile_path"`

	// ExtraIncludePath contains an extra include path that must be
	// used to compile this source file.
	// This is mainly used for source files that comes from old-style libraries
	// (Arduino IDE <1.5) requiring an extra include path to the "utility" folder.
	ExtraIncludePath *paths.Path `json:"extra_include_path,omitempty"`
}

func (f *sourceFile) String() string {
	return fmt.Sprintf("SourcePath:%s SourceRoot:%s BuildRoot:%s ExtraInclude:%s",
		f.SourcePath, f.ObjectPath, f.DepfilePath, f.ExtraIncludePath)
}

// Equals checks if a sourceFile is equal to another.
func (f *sourceFile) Equals(g *sourceFile) bool {
	return f.SourcePath.EqualsTo(g.SourcePath) &&
		f.ObjectPath.EqualsTo(g.ObjectPath) &&
		f.DepfilePath.EqualsTo(g.DepfilePath) &&
		((f.ExtraIncludePath == nil && g.ExtraIncludePath == nil) ||
			(f.ExtraIncludePath != nil && g.ExtraIncludePath != nil && f.ExtraIncludePath.EqualsTo(g.ExtraIncludePath)))
}

// makeSourceFile create a sourceFile object for the given source file path.
// The given sourceFilePath can be absolute, or relative within the sourceRoot root folder.
func makeSourceFile(sourceRoot, buildRoot, sourceFilePath *paths.Path, extraIncludePaths ...*paths.Path) (*sourceFile, error) {
	if len(extraIncludePaths) > 1 {
		panic("only one extra include path allowed")
	}
	var extraIncludePath *paths.Path
	if len(extraIncludePaths) > 0 {
		extraIncludePath = extraIncludePaths[0]
	}

	if sourceFilePath.IsAbs() {
		var err error
		sourceFilePath, err = sourceRoot.RelTo(sourceFilePath)
		if err != nil {
			return nil, err
		}
	}
	res := &sourceFile{
		SourcePath:       sourceRoot.JoinPath(sourceFilePath),
		ObjectPath:       buildRoot.Join(sourceFilePath.String() + ".o"),
		DepfilePath:      buildRoot.Join(sourceFilePath.String() + ".d"),
		ExtraIncludePath: extraIncludePath,
	}
	return res, nil
}

// ObjFileIsUpToDate checks if the compile object file is up to date.
func (f *sourceFile) ObjFileIsUpToDate() (unchanged bool, err error) {
	return utils.ObjFileIsUpToDate(f.SourcePath, f.ObjectPath, f.DepfilePath)
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
