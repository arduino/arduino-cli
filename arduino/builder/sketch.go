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
	"fmt"
	"regexp"
	"strings"

	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"

	"github.com/pkg/errors"
)

var (
	includesArduinoH = regexp.MustCompile(`(?m)^\s*#\s*include\s*[<\"]Arduino\.h[>\"]`)
	tr               = i18n.Tr
)

// QuoteCppString returns the given string as a quoted string for use with the C
// preprocessor. This adds double quotes around it and escapes any
// double quotes and backslashes in the string.
func QuoteCppString(str string) string {
	str = strings.Replace(str, "\\", "\\\\", -1)
	str = strings.Replace(str, "\"", "\\\"", -1)
	return "\"" + str + "\""
}

// SketchSaveItemCpp saves a preprocessed .cpp sketch file on disk
func SketchSaveItemCpp(path *paths.Path, contents []byte, destPath *paths.Path) error {
	sketchName := path.Base()
	if err := destPath.MkdirAll(); err != nil {
		return errors.Wrap(err, tr("unable to create a folder to save the sketch"))
	}

	destFile := destPath.Join(fmt.Sprintf("%s.cpp", sketchName))

	if err := destFile.WriteFile(contents); err != nil {
		return errors.Wrap(err, tr("unable to save the sketch on disk"))
	}

	return nil
}

// SketchMergeSources merges all the source files included in a sketch
func SketchMergeSources(sk *sketch.Sketch, overrides map[string]string) (int, string, error) {
	lineOffset := 0
	mergedSource := ""

	getSource := func(f *paths.Path) (string, error) {
		path, err := sk.FullPath.RelTo(f)
		if err != nil {
			return "", errors.Wrap(err, tr("unable to compute relative path to the sketch for the item"))
		}
		if override, ok := overrides[path.String()]; ok {
			return override, nil
		}
		data, err := f.ReadFile()
		if err != nil {
			return "", fmt.Errorf(tr("reading file %[1]s: %[2]s"), f, err)
		}
		return string(data), nil
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

	mergedSource += "#line 1 " + QuoteCppString(sk.MainFile.String()) + "\n"
	mergedSource += mainSrc + "\n"
	lineOffset++

	for _, file := range sk.OtherSketchFiles {
		src, err := getSource(file)
		if err != nil {
			return 0, "", err
		}
		mergedSource += "#line 1 " + QuoteCppString(file.String()) + "\n"
		mergedSource += src + "\n"
	}

	return lineOffset, mergedSource, nil
}

// SketchCopyAdditionalFiles copies the additional files for a sketch to the
// specified destination directory.
func SketchCopyAdditionalFiles(sketch *sketch.Sketch, destPath *paths.Path, overrides map[string]string) error {
	if err := destPath.MkdirAll(); err != nil {
		return errors.Wrap(err, tr("unable to create a folder to save the sketch files"))
	}

	for _, file := range sketch.AdditionalFiles {
		relpath, err := sketch.FullPath.RelTo(file)
		if err != nil {
			return errors.Wrap(err, tr("unable to compute relative path to the sketch for the item"))
		}

		targetPath := destPath.JoinPath(relpath)
		// create the directory containing the target
		if err = targetPath.Parent().MkdirAll(); err != nil {
			return errors.Wrap(err, tr("unable to create the folder containing the item"))
		}

		var sourceBytes []byte
		if override, ok := overrides[relpath.String()]; ok {
			// use override source
			sourceBytes = []byte(override)
		} else {
			// read the source file
			s, err := file.ReadFile()
			if err != nil {
				return errors.Wrap(err, tr("unable to read contents of the source item"))
			}
			sourceBytes = s
		}

		// tag each addtional file with the filename of the source it was copied from
		sourceBytes = append([]byte("#line 1 "+QuoteCppString(file.String())+"\n"), sourceBytes...)

		err = writeIfDifferent(sourceBytes, targetPath)
		if err != nil {
			return errors.Wrap(err, tr("unable to write to destination file"))
		}
	}

	return nil
}

func writeIfDifferent(source []byte, destPath *paths.Path) error {
	// Check whether the destination file exists
	if destPath.NotExist() {
		// Write directly
		return destPath.WriteFile(source)
	}

	// Read the destination file if it exists
	existingBytes, err := destPath.ReadFile()
	if err != nil {
		return errors.Wrap(err, tr("unable to read contents of the destination item"))
	}

	// Overwrite if contents are different
	if !bytes.Equal(existingBytes, source) {
		return destPath.WriteFile(source)
	}

	// Source and destination are the same, don't write anything
	return nil
}

// SetupBuildProperties adds the build properties related to the sketch to the
// default board build properties map.
func SetupBuildProperties(boardBuildProperties *properties.Map, buildPath *paths.Path, sketch *sketch.Sketch, optimizeForDebug bool) *properties.Map {
	buildProperties := properties.NewMap()
	buildProperties.Merge(boardBuildProperties)

	if buildPath != nil {
		buildProperties.SetPath("build.path", buildPath)
	}
	if sketch != nil {
		buildProperties.Set("build.project_name", sketch.MainFile.Base())
		buildProperties.SetPath("build.source.path", sketch.FullPath)
	}
	if optimizeForDebug {
		if debugFlags, ok := buildProperties.GetOk("compiler.optimization_flags.debug"); ok {
			buildProperties.Set("compiler.optimization_flags", debugFlags)
		}
	} else {
		if releaseFlags, ok := buildProperties.GetOk("compiler.optimization_flags.release"); ok {
			buildProperties.Set("compiler.optimization_flags", releaseFlags)
		}
	}

	return buildProperties
}
