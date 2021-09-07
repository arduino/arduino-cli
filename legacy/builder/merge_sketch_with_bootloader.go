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
	"math"
	"strconv"
	"strings"

	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	"github.com/arduino/go-paths-helper"
	"github.com/marcinbor85/gohex"
	"github.com/pkg/errors"
)

type MergeSketchWithBootloader struct{}

func (s *MergeSketchWithBootloader) Run(ctx *types.Context) error {
	if ctx.OnlyUpdateCompilationDatabase {
		return nil
	}

	buildProperties := ctx.BuildProperties
	if !buildProperties.ContainsKey(constants.BUILD_PROPERTIES_BOOTLOADER_NOBLINK) && !buildProperties.ContainsKey(constants.BUILD_PROPERTIES_BOOTLOADER_FILE) {
		return nil
	}

	buildPath := ctx.BuildPath
	sketch := ctx.Sketch
	sketchFileName := sketch.MainFile.Base()

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
		utils.LogIfVerbose(constants.LOG_LEVEL_WARN, tr("Bootloader file specified but missing: {0}"), bootloaderPath)
		return nil
	}

	mergedSketchPath := builtSketchPath.Parent().Join(sketchFileName + ".with_bootloader.hex")

	// Ignore merger errors for the first iteration
	maximumBinSize := 16000000
	if uploadMaxSize, ok := ctx.BuildProperties.GetOk(constants.PROPERTY_UPLOAD_MAX_SIZE); ok {
		maximumBinSize, _ = strconv.Atoi(uploadMaxSize)
		maximumBinSize *= 2
	}
	err := merge(builtSketchPath, bootloaderPath, mergedSketchPath, maximumBinSize)
	if err != nil {
		utils.LogIfVerbose(constants.LOG_LEVEL_INFO, err.Error())
	}

	return nil
}

func merge(builtSketchPath, bootloaderPath, mergedSketchPath *paths.Path, maximumBinSize int) error {
	if bootloaderPath.Ext() == ".bin" {
		bootloaderPath = paths.New(strings.TrimSuffix(bootloaderPath.String(), ".bin") + ".hex")
	}

	memBoot := gohex.NewMemory()
	if bootFile, err := bootloaderPath.Open(); err == nil {
		defer bootFile.Close()
		if err := memBoot.ParseIntelHex(bootFile); err != nil {
			return errors.New(bootFile.Name() + " " + err.Error())
		}
	} else {
		return err
	}

	memSketch := gohex.NewMemory()
	if buildFile, err := builtSketchPath.Open(); err == nil {
		defer buildFile.Close()
		if err := memSketch.ParseIntelHex(buildFile); err != nil {
			return errors.New(buildFile.Name() + " " + err.Error())
		}
	} else {
		return err
	}

	memMerged := gohex.NewMemory()
	initialAddress := uint32(math.MaxUint32)
	lastAddress := uint32(0)

	for _, segment := range memBoot.GetDataSegments() {
		if err := memMerged.AddBinary(segment.Address, segment.Data); err != nil {
			continue
		}
		if segment.Address < initialAddress {
			initialAddress = segment.Address
		}
		if segment.Address+uint32(len(segment.Data)) > lastAddress {
			lastAddress = segment.Address + uint32(len(segment.Data))
		}
	}
	for _, segment := range memSketch.GetDataSegments() {
		if err := memMerged.AddBinary(segment.Address, segment.Data); err != nil {
			continue
		}
		if segment.Address < initialAddress {
			initialAddress = segment.Address
		}
		if segment.Address+uint32(len(segment.Data)) > lastAddress {
			lastAddress = segment.Address + uint32(len(segment.Data))
		}
	}

	if mergeFile, err := mergedSketchPath.Create(); err == nil {
		defer mergeFile.Close()
		memMerged.DumpIntelHex(mergeFile, 16)
	} else {
		return err
	}

	// Write out a .bin if the addresses doesn't go too far away from origin
	// (and consequently produce a very large bin)
	size := lastAddress - initialAddress
	if size > uint32(maximumBinSize) {
		return nil
	}
	mergedSketchPathBin := paths.New(strings.TrimSuffix(mergedSketchPath.String(), ".hex") + ".bin")
	data := memMerged.ToBinary(initialAddress, size, 0xFF)
	return mergedSketchPathBin.WriteFile(data)
}
