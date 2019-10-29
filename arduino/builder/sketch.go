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

package builder

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/arduino/arduino-cli/arduino/globals"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"

	"github.com/pkg/errors"
)

// As currently implemented on Linux,
// the maximum number of symbolic links that will be followed while resolving a pathname is 40
const maxFileSystemDepth = 40

var includesArduinoH = regexp.MustCompile(`(?m)^\s*#\s*include\s*[<\"]Arduino\.h[>\"]`)

// QuoteCppString returns the given string as a quoted string for use with the C
// preprocessor. This adds double quotes around it and escapes any
// double quotes and backslashes in the string.
func QuoteCppString(str string) string {
	str = strings.Replace(str, "\\", "\\\\", -1)
	str = strings.Replace(str, "\"", "\\\"", -1)
	return "\"" + str + "\""
}

// SketchSaveItemCpp saves a preprocessed .cpp sketch file on disk
func SketchSaveItemCpp(item *sketch.Item, destPath string) error {

	sketchName := filepath.Base(item.Path)

	if err := os.MkdirAll(destPath, os.FileMode(0755)); err != nil {
		return errors.Wrap(err, "unable to create a folder to save the sketch")
	}

	destFile := filepath.Join(destPath, sketchName+".cpp")

	if err := ioutil.WriteFile(destFile, item.Source, os.FileMode(0644)); err != nil {
		return errors.Wrap(err, "unable to save the sketch on disk")
	}

	return nil
}

// simpleLocalWalk locally replaces filepath.Walk and/but goes through symlinks
func simpleLocalWalk(root string, maxDepth int, walkFn func(path string, info os.FileInfo, err error) error) error {

	info, err := os.Stat(root)

	if err != nil {
		return walkFn(root, nil, err)
	}

	err = walkFn(root, info, err)
	if err == filepath.SkipDir {
		return nil
	}

	if info.IsDir() {
		if maxDepth <= 0 {
			return walkFn(root, info, errors.New("Filesystem bottom is too deep (directory recursion or filesystem really deep): "+root))
		}
		maxDepth--
		files, err := ioutil.ReadDir(root)
		if err == nil {
			for _, file := range files {
				err = simpleLocalWalk(root+string(os.PathSeparator)+file.Name(), maxDepth, walkFn)
				if err == filepath.SkipDir {
					return nil
				}
			}
		}
	}

	return nil
}

// SketchLoad collects all the files composing a sketch.
// The parameter `sketchPath` holds a path pointing to a single sketch file or a sketch folder,
// the path must be absolute.
func SketchLoad(sketchPath, buildPath string) (*sketch.Sketch, error) {
	stat, err := os.Stat(sketchPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to stat Sketch location")
	}

	var sketchFolder, mainSketchFile string

	// if a sketch folder was passed, save the parent and point sketchPath to the main .ino file
	if stat.IsDir() {
		sketchFolder = sketchPath
		mainSketchFile = filepath.Join(sketchPath, stat.Name()+".ino")
		// in the case a dir was passed, ensure the main file exists and is readable
		f, err := os.Open(mainSketchFile)
		if err != nil {
			return nil, errors.Wrap(err, "unable to find the main sketch file")
		}
		f.Close()
		// ensure it is not a directory
		info, err := os.Stat(mainSketchFile)
		if err != nil {
			return nil, errors.Wrap(err, "unable to check the main sketch file")
		}
		if info.IsDir() {
			return nil, errors.Wrap(errors.New(mainSketchFile), "sketch must not be a directory")
		}
	} else {
		sketchFolder = filepath.Dir(sketchPath)
		mainSketchFile = sketchPath
	}

	// collect all the sketch files
	var files []string
	err = simpleLocalWalk(sketchFolder, maxFileSystemDepth, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			feedback.Errorf("Error during sketch processing: %v", err)
			os.Exit(errorcodes.ErrGeneric)
		}

		// ignore hidden files and skip hidden directories
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// skip legacy SCM directories
		if info.IsDir() && strings.HasPrefix(info.Name(), "CVS") || strings.HasPrefix(info.Name(), "RCS") {
			return filepath.SkipDir
		}

		// ignore directory entries
		if info.IsDir() {
			return nil
		}

		// ignore if file extension doesn't match
		ext := strings.ToLower(filepath.Ext(path))
		_, isMain := globals.MainFileValidExtensions[ext]
		_, isAdditional := globals.AdditionalFileValidExtensions[ext]
		if !(isMain || isAdditional) {
			return nil
		}

		// check if file is readable
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		f.Close()

		// collect the file
		files = append(files, path)

		// done
		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "there was an error while collecting the sketch files")
	}

	return sketch.New(sketchFolder, mainSketchFile, buildPath, files)
}

// SketchMergeSources merges all the source files included in a sketch
func SketchMergeSources(sketch *sketch.Sketch) (int, string) {
	lineOffset := 0
	mergedSource := ""

	// add Arduino.h inclusion directive if missing
	if !includesArduinoH.MatchString(sketch.MainFile.GetSourceStr()) {
		mergedSource += "#include <Arduino.h>\n"
		lineOffset++
	}

	mergedSource += "#line 1 " + QuoteCppString(sketch.MainFile.Path) + "\n"
	mergedSource += sketch.MainFile.GetSourceStr() + "\n"
	lineOffset++

	for _, item := range sketch.OtherSketchFiles {
		mergedSource += "#line 1 " + QuoteCppString(item.Path) + "\n"
		mergedSource += item.GetSourceStr() + "\n"
	}

	return lineOffset, mergedSource
}

// SketchCopyAdditionalFiles copies the additional files for a sketch to the
// specified destination directory.
func SketchCopyAdditionalFiles(sketch *sketch.Sketch, destPath string) error {
	if err := os.MkdirAll(destPath, os.FileMode(0755)); err != nil {
		return errors.Wrap(err, "unable to create a folder to save the sketch files")
	}

	for _, item := range sketch.AdditionalFiles {
		relpath, err := filepath.Rel(sketch.LocationPath, item.Path)
		if err != nil {
			return errors.Wrap(err, "unable to compute relative path to the sketch for the item")
		}

		targetPath := filepath.Join(destPath, relpath)
		// create the directory containing the target
		if err = os.MkdirAll(filepath.Dir(targetPath), os.FileMode(0755)); err != nil {
			return errors.Wrap(err, "unable to create the folder containing the item")
		}

		err = writeIfDifferent(item.Path, targetPath)
		if err != nil {
			return errors.Wrap(err, "unable to write to destination file")
		}
	}

	return nil
}

func writeIfDifferent(sourcePath, destPath string) error {
	// read the source file
	newbytes, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		return errors.Wrap(err, "unable to read contents of the source item")
	}

	// check whether the destination file exists
	_, err = os.Stat(destPath)
	if os.IsNotExist(err) {
		// write directly
		return ioutil.WriteFile(destPath, newbytes, os.FileMode(0644))
	}

	// read the destination file if it ex
	existingBytes, err := ioutil.ReadFile(destPath)
	if err != nil {
		return errors.Wrap(err, "unable to read contents of the destination item")
	}

	// overwrite if contents are different
	if bytes.Compare(existingBytes, newbytes) != 0 {
		return ioutil.WriteFile(destPath, newbytes, os.FileMode(0644))
	}

	// source and destination are the same, don't write anything
	return nil
}
