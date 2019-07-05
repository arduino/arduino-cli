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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-cli/arduino/globals"
	"github.com/arduino/arduino-cli/arduino/sketch"

	"github.com/pkg/errors"
)

// SaveSketchItemCpp saves a preprocessed .cpp sketch file on disk
func SaveSketchItemCpp(item *sketch.Item, buildPath string) error {

	sketchName := filepath.Base(item.Path)

	if err := os.MkdirAll(buildPath, os.FileMode(0755)); err != nil {
		return errors.Wrap(err, "unable to create a folder to save the sketch")
	}

	destFile := filepath.Join(buildPath, sketchName+".cpp")

	if err := ioutil.WriteFile(destFile, item.Source, os.FileMode(0644)); err != nil {
		return errors.Wrap(err, "unable to save the sketch on disk")
	}

	return nil
}

// LoadSketch collects all the files composing a sketch.
// The parameter `sketchPath` holds a path pointing to a single sketch file or a sketch folder,
// the path must be absolute.
func LoadSketch(sketchPath, buildPath string) (*sketch.Sketch, error) {
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
	} else {
		sketchFolder = filepath.Dir(sketchPath)
		mainSketchFile = sketchPath
	}

	// collect all the sketch files
	var files []string
	err = filepath.Walk(sketchFolder, func(path string, info os.FileInfo, err error) error {
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
