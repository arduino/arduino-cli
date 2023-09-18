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
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/arduino/arduino-cli/arduino/builder/cpp"
	"github.com/arduino/arduino-cli/arduino/builder/utils"
	"github.com/arduino/arduino-cli/i18n"
	f "github.com/arduino/arduino-cli/internal/algorithms"
	"github.com/arduino/go-paths-helper"
	"github.com/marcinbor85/gohex"

	"github.com/pkg/errors"
)

var (
	includesArduinoH = regexp.MustCompile(`(?m)^\s*#\s*include\s*[<\"]Arduino\.h[>\"]`)
	tr               = i18n.Tr
)

// prepareSketchBuildPath copies the sketch source files in the build path.
// The .ino files are merged together to create a .cpp file (by the way, the
// .cpp file still needs to be Arduino-preprocessed to compile).
func (b *Builder) prepareSketchBuildPath() error {
	if err := b.sketchBuildPath.MkdirAll(); err != nil {
		return errors.Wrap(err, tr("unable to create a folder to save the sketch"))
	}

	offset, mergedSource, err := b.sketchMergeSources(b.sourceOverrides)
	if err != nil {
		return err
	}

	destFile := b.sketchBuildPath.Join(b.sketch.MainFile.Base() + ".cpp")
	if err := destFile.WriteFile([]byte(mergedSource)); err != nil {
		return err
	}

	if err := b.sketchCopyAdditionalFiles(b.sketchBuildPath, b.sourceOverrides); err != nil {
		return err
	}

	b.lineOffset = offset

	return nil
}

// sketchMergeSources merges all the .ino source files included in a sketch to produce
// a single .cpp file.
func (b *Builder) sketchMergeSources(overrides map[string]string) (int, string, error) {
	lineOffset := 0
	mergedSource := ""

	getSource := func(f *paths.Path) (string, error) {
		path, err := b.sketch.FullPath.RelTo(f)
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
	mainSrc, err := getSource(b.sketch.MainFile)
	if err != nil {
		return 0, "", err
	}
	if !includesArduinoH.MatchString(mainSrc) {
		mergedSource += "#include <Arduino.h>\n"
		lineOffset++
	}

	mergedSource += "#line 1 " + cpp.QuoteString(b.sketch.MainFile.String()) + "\n"
	mergedSource += mainSrc + "\n"
	lineOffset++

	for _, file := range b.sketch.OtherSketchFiles {
		src, err := getSource(file)
		if err != nil {
			return 0, "", err
		}
		mergedSource += "#line 1 " + cpp.QuoteString(file.String()) + "\n"
		mergedSource += src + "\n"
	}

	return lineOffset, mergedSource, nil
}

// sketchCopyAdditionalFiles copies the additional files for a sketch to the
// specified destination directory.
func (b *Builder) sketchCopyAdditionalFiles(buildPath *paths.Path, overrides map[string]string) error {
	for _, file := range b.sketch.AdditionalFiles {
		relpath, err := b.sketch.FullPath.RelTo(file)
		if err != nil {
			return errors.Wrap(err, tr("unable to compute relative path to the sketch for the item"))
		}

		targetPath := buildPath.JoinPath(relpath)
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
		sourceBytes = append([]byte("#line 1 "+cpp.QuoteString(file.String())+"\n"), sourceBytes...)

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

// BuildSketch fixdoc
func (b *Builder) BuildSketch(includesFolders paths.PathList) error {
	includes := f.Map(includesFolders.AsStrings(), cpp.WrapWithHyphenI)

	if err := b.sketchBuildPath.MkdirAll(); err != nil {
		return errors.WithStack(err)
	}

	sketchObjectFiles, err := utils.CompileFiles(
		b.sketchBuildPath, b.sketchBuildPath, b.buildProperties, includes,
		b.onlyUpdateCompilationDatabase,
		b.compilationDatabase,
		b.jobs,
		b.logger,
		b.Progress,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	// The "src/" subdirectory of a sketch is compiled recursively
	sketchSrcPath := b.sketchBuildPath.Join("src")
	if sketchSrcPath.IsDir() {
		srcObjectFiles, err := utils.CompileFilesRecursive(
			sketchSrcPath, sketchSrcPath, b.buildProperties, includes,
			b.onlyUpdateCompilationDatabase,
			b.compilationDatabase,
			b.jobs,
			b.logger,
			b.Progress,
		)
		if err != nil {
			return errors.WithStack(err)
		}
		sketchObjectFiles.AddAll(srcObjectFiles)
	}

	b.buildArtifacts.sketchObjectFiles = sketchObjectFiles
	return nil
}

// MergeSketchWithBootloader fixdoc
func (b *Builder) MergeSketchWithBootloader() error {
	if b.onlyUpdateCompilationDatabase {
		return nil
	}

	if !b.buildProperties.ContainsKey("bootloader.noblink") && !b.buildProperties.ContainsKey("bootloader.file") {
		return nil
	}

	sketchFileName := b.sketch.MainFile.Base()
	sketchInBuildPath := b.buildPath.Join(sketchFileName + ".hex")
	sketchInSubfolder := b.buildPath.Join("sketch", sketchFileName+".hex")

	var builtSketchPath *paths.Path
	if sketchInBuildPath.Exist() {
		builtSketchPath = sketchInBuildPath
	} else if sketchInSubfolder.Exist() {
		builtSketchPath = sketchInSubfolder
	} else {
		return nil
	}

	bootloader := ""
	if bootloaderNoBlink, ok := b.buildProperties.GetOk("bootloader.noblink"); ok {
		bootloader = bootloaderNoBlink
	} else {
		bootloader = b.buildProperties.Get("bootloader.file")
	}
	bootloader = b.buildProperties.ExpandPropsInString(bootloader)

	bootloaderPath := b.buildProperties.GetPath("runtime.platform.path").Join("bootloaders", bootloader)
	if bootloaderPath.NotExist() {
		if b.logger.Verbose() {
			b.logger.Warn(tr("Bootloader file specified but missing: %[1]s", bootloaderPath))
		}
		return nil
	}

	mergedSketchPath := builtSketchPath.Parent().Join(sketchFileName + ".with_bootloader.hex")

	// Ignore merger errors for the first iteration
	maximumBinSize := 16000000
	if uploadMaxSize, ok := b.buildProperties.GetOk("upload.maximum_size"); ok {
		maximumBinSize, _ = strconv.Atoi(uploadMaxSize)
		maximumBinSize *= 2
	}
	err := merge(builtSketchPath, bootloaderPath, mergedSketchPath, maximumBinSize)
	if err != nil && b.logger.Verbose() {
		b.logger.Info(err.Error())
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
