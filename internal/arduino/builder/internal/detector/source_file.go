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

	"os"

	"github.com/arduino/arduino-cli/internal/arduino/builder/cpp"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

type sourceFile struct {
	// SourcePath is the path to the source file
	SourcePath *paths.Path `json:"source_path"`

	// DepfilePath is the path to the dependency file that will be generated
	DepfilePath *paths.Path `json:"depfile_path"`

	// ExtraIncludePath contains an extra include path that must be
	// used to compile this source file.
	// This is mainly used for source files that comes from old-style libraries
	// (Arduino IDE <1.5) requiring an extra include path to the "utility" folder.
	ExtraIncludePath *paths.Path `json:"extra_include_path,omitempty"`
}

func (f *sourceFile) String() string {
	return fmt.Sprintf("%s -> dep:%s (ExtraInclude:%s)",
		f.SourcePath, f.DepfilePath, f.ExtraIncludePath)
}

// Equals checks if a sourceFile is equal to another.
func (f *sourceFile) Equals(g sourceFile) bool {
	return f.SourcePath.EqualsTo(g.SourcePath) &&
		f.DepfilePath.EqualsTo(g.DepfilePath) &&
		((f.ExtraIncludePath == nil && g.ExtraIncludePath == nil) ||
			(f.ExtraIncludePath != nil && g.ExtraIncludePath != nil && f.ExtraIncludePath.EqualsTo(g.ExtraIncludePath)))
}

// PrepareBuildPath ensures that the directory for the dependency file exists.
func (f *sourceFile) PrepareBuildPath() error {
	if f.DepfilePath != nil {
		return f.DepfilePath.Parent().MkdirAll()
	}
	return nil
}

// ObjFileIsUpToDate checks if the compile object file is up to date.
func (f *sourceFile) ObjFileIsUpToDate(log *logrus.Entry) (unchanged bool, err error) {
	if f.DepfilePath == nil {
		log.Tracef("[LD] COMPILE-CHECK: REBUILD %v: object file or dependency file not provided", f.SourcePath)
		return false, nil
	}

	sourceFile := f.SourcePath.Clean()
	sourceFileStat, err := sourceFile.Stat()
	if err != nil {
		log.Tracef("[LD] COMPILE-CHECK: REBUILD %v: Could not stat source file: %s", f.SourcePath, err)
		return false, err
	}
	dependencyFile := f.DepfilePath.Clean()
	dependencyFileStat, err := dependencyFile.Stat()
	if err != nil {
		if os.IsNotExist(err) {
			log.Tracef("[LD] COMPILE-CHECK: REBUILD %v: Dependency file not found: %v", f.SourcePath, dependencyFile)
			return false, nil
		}
		log.Tracef("[LD] COMPILE-CHECK: REBUILD %v: Could not stat dependency file: %s", f.SourcePath, err)
		return false, err
	}
	if sourceFileStat.ModTime().After(dependencyFileStat.ModTime()) {
		log.Tracef("[LD] COMPILE-CHECK: REBUILD %v: %v newer than %v", f.SourcePath, sourceFile, dependencyFile)
		return false, nil
	}
	deps, err := cpp.ReadDepFile(dependencyFile)
	if err != nil {
		log.Tracef("[LD] COMPILE-CHECK: REBUILD %v: Could not read dependency file: %s", f.SourcePath, dependencyFile)
		return false, err
	}
	if len(deps.Dependencies) == 0 {
		return true, nil
	}
	if deps.Dependencies[0] != sourceFile.String() {
		log.Tracef("[LD] COMPILE-CHECK: REBUILD %v: Depfile is about different source file: %v (expected %v)", f.SourcePath, deps.Dependencies[0], sourceFile)
		return false, nil
	}
	for _, dep := range deps.Dependencies[1:] {
		depStat, err := os.Stat(dep)
		if os.IsNotExist(err) {
			log.Tracef("[LD] COMPILE-CHECK: REBUILD %v: Not found: %v", f.SourcePath, dep)
			return false, nil
		}
		if err != nil {
			logrus.WithError(err).Tracef("[LD] COMPILE-CHECK: REBUILD %v: Failed to read: %v", f.SourcePath, dep)
			return false, nil
		}
		if depStat.ModTime().After(dependencyFileStat.ModTime()) {
			log.Tracef("[LD] COMPILE-CHECK: REBUILD %v: %v newer than %v", f.SourcePath, dep, dependencyFile)
			return false, nil
		}
	}
	log.Tracef("[LD] COMPILE-CHECK: REUSE %v Up-to-date", f.SourcePath)
	return true, nil
}

// uniqueSourceFileQueue is a queue of source files that does not allow duplicates.
type uniqueSourceFileQueue []sourceFile

// Push adds a source file to the queue if it is not already present.
func (queue *uniqueSourceFileQueue) Push(value sourceFile) {
	if !queue.Contains(value) {
		logrus.Tracef("[LD] QUEUE: Added %s", value.SourcePath)
		*queue = append(*queue, value)
	}
}

// Contains checks if the queue Contains a source file.
func (queue uniqueSourceFileQueue) Contains(target sourceFile) bool {
	return slices.ContainsFunc(queue, target.Equals)
}

// Pop removes and returns the first element of the queue.
func (queue *uniqueSourceFileQueue) Pop() sourceFile {
	old := *queue
	x := old[0]
	*queue = old[1:]
	return x
}

// Empty returns true if the queue is Empty.
func (queue uniqueSourceFileQueue) Empty() bool {
	return len(queue) == 0
}
