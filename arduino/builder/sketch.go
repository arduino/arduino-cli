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
func SketchSaveItemCpp(path string, contents []byte, destPath string) error {

	sketchName := filepath.Base(path)

	if err := os.MkdirAll(destPath, os.FileMode(0755)); err != nil {
		return errors.Wrap(err, "unable to create a folder to save the sketch")
	}

	destFile := filepath.Join(destPath, sketchName+".cpp")

	if err := ioutil.WriteFile(destFile, contents, os.FileMode(0644)); err != nil {
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

	// if a sketch folder was passed, save the parent and point sketchPath to the main sketch file
	if stat.IsDir() {
		sketchFolder = sketchPath
		// allowed extensions are .ino and .pde (but not both)
		for extension := range globals.MainFileValidExtensions {
			candidateSketchFile := filepath.Join(sketchPath, stat.Name()+extension)
			if _, err := os.Stat(candidateSketchFile); !os.IsNotExist(err) {
				if mainSketchFile == "" {
					mainSketchFile = candidateSketchFile
				} else {
					return nil, errors.Errorf("multiple main sketch files found (%v,%v)",
						filepath.Base(mainSketchFile),
						filepath.Base(candidateSketchFile))
				}
			}
		}

		// check main file was found
		if mainSketchFile == "" {
			return nil, errors.Errorf("unable to find a sketch file in directory %v", sketchFolder)
		}

		// check main file is readable
		f, err := os.Open(mainSketchFile)
		if err != nil {
			return nil, errors.Wrap(err, "unable to open the main sketch file")
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
	rootVisited := false
	err = simpleLocalWalk(sketchFolder, maxFileSystemDepth, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			feedback.Errorf("Error during sketch processing: %v", err)
			os.Exit(errorcodes.ErrGeneric)
		}

		if info.IsDir() {
			// Filters in this if-block are NOT applied to the sketch folder itself.
			// Since the sketch folder is the first one processed by simpleLocalWalk,
			// we can set the `rootVisited` guard to exclude it.
			if rootVisited {
				// skip hidden folders
				if strings.HasPrefix(info.Name(), ".") {
					return filepath.SkipDir
				}

				// skip legacy SCM directories
				if info.Name() == "CVS" || info.Name() == "RCS" {
					return filepath.SkipDir
				}
			} else {
				rootVisited = true
			}

			// ignore (don't skip) directory
			return nil
		}

		// ignore hidden files
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// ignore if file extension doesn't match
		ext := filepath.Ext(path)
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
func SketchMergeSources(sk *sketch.Sketch, overrides map[string]string) (int, string, error) {
	lineOffset := 0
	mergedSource := ""

	getSource := func(i *sketch.Item) (string, error) {
		path, err := filepath.Rel(sk.LocationPath, i.Path)
		if err != nil {
			return "", errors.Wrap(err, "unable to compute relative path to the sketch for the item")
		}
		if override, ok := overrides[path]; ok {
			return override, nil
		}
		return i.GetSourceStr()
	}

	// add Arduino.h inclusion directive if missing
	mainSrc, err := getSource(sk.MainFile)
	if err != nil {
		return 0, "", err
	}
	if !includesArduinoH.MatchString(mainSrc) {
		mergedSource += "#include <Arduino.h>\n"
		lineOffset++
	}

	mergedSource += "#line 1 " + QuoteCppString(sk.MainFile.Path) + "\n"
	mergedSource += mainSrc + "\n"
	lineOffset++

	for _, item := range sk.OtherSketchFiles {
		src, err := getSource(item)
		if err != nil {
			return 0, "", err
		}
		mergedSource += "#line 1 " + QuoteCppString(item.Path) + "\n"
		mergedSource += src + "\n"
	}

	return lineOffset, mergedSource, nil
}

// SketchCopyAdditionalFiles copies the additional files for a sketch to the
// specified destination directory.
func SketchCopyAdditionalFiles(sketch *sketch.Sketch, destPath string, overrides map[string]string) error {
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

		var sourceBytes []byte
		if override, ok := overrides[relpath]; ok {
			// use override source
			sourceBytes = []byte(override)
		} else {
			// read the source file
			s, err := item.GetSourceBytes()
			if err != nil {
				return errors.Wrap(err, "unable to read contents of the source item")
			}
			sourceBytes = s
		}

		// tag each addtional file with the filename of the source it was copied from
		sourceBytes = append([]byte("#line 1 "+QuoteCppString(item.Path)+"\n"), sourceBytes...)

		err = writeIfDifferent(sourceBytes, targetPath)
		if err != nil {
			return errors.Wrap(err, "unable to write to destination file")
		}
	}

	return nil
}

func writeIfDifferent(source []byte, destPath string) error {
	// check whether the destination file exists
	_, err := os.Stat(destPath)
	if os.IsNotExist(err) {
		// write directly
		return ioutil.WriteFile(destPath, source, os.FileMode(0644))
	}

	// read the destination file if it ex
	existingBytes, err := ioutil.ReadFile(destPath)
	if err != nil {
		return errors.Wrap(err, "unable to read contents of the destination item")
	}

	// overwrite if contents are different
	if bytes.Compare(existingBytes, source) != 0 {
		return ioutil.WriteFile(destPath, source, os.FileMode(0644))
	}

	// source and destination are the same, don't write anything
	return nil
}
