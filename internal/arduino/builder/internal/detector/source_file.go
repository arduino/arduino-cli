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
func (f *sourceFile) Equals(g *sourceFile) bool {
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
func (f *sourceFile) ObjFileIsUpToDate() (unchanged bool, err error) {
	logrus.Debugf("Checking previous results for %v (dep = %v)", f.SourcePath, f.DepfilePath)
	if f.DepfilePath == nil {
		logrus.Debugf("  Object file or dependency file not provided")
		return false, nil
	}

	sourceFile := f.SourcePath.Clean()
	sourceFileStat, err := sourceFile.Stat()
	if err != nil {
		logrus.Debugf("  Could not stat source file: %s", err)
		return false, err
	}
	dependencyFile := f.DepfilePath.Clean()
	dependencyFileStat, err := dependencyFile.Stat()
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Debugf("  Dependency file not found: %v", dependencyFile)
			return false, nil
		}
		logrus.Debugf("  Could not stat dependency file: %s", err)
		return false, err
	}
	if sourceFileStat.ModTime().After(dependencyFileStat.ModTime()) {
		logrus.Debugf("  %v newer than %v", sourceFile, dependencyFile)
		return false, nil
	}
	deps, err := cpp.ReadDepFile(dependencyFile)
	if err != nil {
		logrus.Debugf("  Could not read dependency file: %s", dependencyFile)
		return false, err
	}
	if len(deps.Dependencies) == 0 {
		return true, nil
	}
	if deps.Dependencies[0] != sourceFile.String() {
		logrus.Debugf("  Depfile is about different source file: %v (expected %v)", deps.Dependencies[0], sourceFile)
		return false, nil
	}
	for _, dep := range deps.Dependencies[1:] {
		depStat, err := os.Stat(dep)
		if os.IsNotExist(err) {
			logrus.Debugf("  Not found: %v", dep)
			return false, nil
		}
		if err != nil {
			logrus.WithError(err).Debugf("  Failed to read: %v", dep)
			return false, nil
		}
		if depStat.ModTime().After(dependencyFileStat.ModTime()) {
			logrus.Debugf("  %v newer than %v", dep, dependencyFile)
			return false, nil
		}
	}
	return true, nil
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
