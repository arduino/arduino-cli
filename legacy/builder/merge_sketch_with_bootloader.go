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
	"os"
	"strings"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/i18n"
	"github.com/arduino/arduino-cli/legacy/builder/types"
)

type MergeSketchWithBootloader struct{}

func (s *MergeSketchWithBootloader) Run(ctx *types.Context) error {
	buildProperties := ctx.BuildProperties
	if !buildProperties.ContainsKey(constants.BUILD_PROPERTIES_BOOTLOADER_NOBLINK) && !buildProperties.ContainsKey(constants.BUILD_PROPERTIES_BOOTLOADER_FILE) {
		return nil
	}

	buildPath := ctx.BuildPath
	sketch := ctx.Sketch
	sketchFileName := sketch.MainFile.Name.Base()
	logger := ctx.GetLogger()

	sketchInBuildPath := buildPath.Join(sketchFileName + ".hex")
	sketchInSubfolder := buildPath.Join(constants.FOLDER_SKETCH, sketchFileName+".hex")

	var builtSketchPath *paths.Path
	if sketchInBuildPath.Exist() {
		builtSketchPath = sketchInBuildPath
	} else if sketchInSubfolder.Exist() {
		builtSketchPath = sketchInSubfolder
	} else {
		return nil
	}

	bootloader := constants.EMPTY_STRING
	if bootloaderNoBlink, ok := buildProperties.GetOk(constants.BUILD_PROPERTIES_BOOTLOADER_NOBLINK); ok {
		bootloader = bootloaderNoBlink
	} else {
		bootloader = buildProperties.Get(constants.BUILD_PROPERTIES_BOOTLOADER_FILE)
	}
	bootloader = buildProperties.ExpandPropsInString(bootloader)

	bootloaderPath := buildProperties.GetPath(constants.BUILD_PROPERTIES_RUNTIME_PLATFORM_PATH).Join(constants.FOLDER_BOOTLOADERS, bootloader)
	if bootloaderPath.NotExist() {
		logger.Fprintln(os.Stdout, constants.LOG_LEVEL_WARN, constants.MSG_BOOTLOADER_FILE_MISSING, bootloaderPath)
		return nil
	}

	mergedSketchPath := builtSketchPath.Parent().Join(sketchFileName + ".with_bootloader.hex")

	return merge(builtSketchPath, bootloaderPath, mergedSketchPath)
}

func hexLineOnlyContainsFF(line string) bool {
	//:206FE000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFB1
	if len(line) <= 11 {
		return false
	}
	byteArray := []byte(line)
	for _, char := range byteArray[9:(len(byteArray) - 2)] {
		if char != 'F' {
			return false
		}
	}
	return true
}

func extractActualBootloader(bootloader []string) []string {

	var realBootloader []string

	// skip until we find a line full of FFFFFF (except address and checksum)
	for i, row := range bootloader {
		if hexLineOnlyContainsFF(row) {
			realBootloader = bootloader[i:len(bootloader)]
			break
		}
	}

	// drop all "empty" lines
	for i, row := range realBootloader {
		if !hexLineOnlyContainsFF(row) {
			realBootloader = realBootloader[i:len(realBootloader)]
			break
		}
	}

	if len(realBootloader) == 0 {
		// we didn't find any line full of FFFF, thus it's a standalone bootloader
		realBootloader = bootloader
	}

	return realBootloader
}

func merge(builtSketchPath, bootloaderPath, mergedSketchPath *paths.Path) error {
	sketch, err := builtSketchPath.ReadFileAsLines()
	if err != nil {
		return i18n.WrapError(err)
	}
	sketch = sketch[:len(sketch)-2]

	bootloader, err := bootloaderPath.ReadFileAsLines()
	if err != nil {
		return i18n.WrapError(err)
	}

	realBootloader := extractActualBootloader(bootloader)

	for _, row := range realBootloader {
		sketch = append(sketch, row)
	}

	return mergedSketchPath.WriteFile([]byte(strings.Join(sketch, "\n")))
}
