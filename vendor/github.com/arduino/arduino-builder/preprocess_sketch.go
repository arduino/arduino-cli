/*
 * This file is part of Arduino Builder.
 *
 * Arduino Builder is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2015 Arduino LLC (http://www.arduino.cc/)
 */

package builder

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/arduino/arduino-builder/constants"
	"github.com/arduino/arduino-builder/i18n"
	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/arduino-builder/utils"
	properties "github.com/arduino/go-properties-orderedmap"
)

// ArduinoPreprocessorProperties are the platform properties needed to run arduino-preprocessor
var ArduinoPreprocessorProperties = properties.NewFromHashmap(map[string]string{
	// Ctags
	"tools.arduino-preprocessor.path":     "{runtime.tools.arduino-preprocessor.path}",
	"tools.arduino-preprocessor.cmd.path": "{path}/arduino-preprocessor",
	"tools.arduino-preprocessor.pattern":  `"{cmd.path}" "{source_file}" "{codecomplete}" -- -std=gnu++11`,

	"preproc.macros.flags": "-w -x c++ -E -CC",
})

type PreprocessSketchArduino struct{}

func (s *PreprocessSketchArduino) Run(ctx *types.Context) error {
	sourceFile := ctx.SketchBuildPath.Join(ctx.Sketch.MainFile.Name.Base() + ".cpp")
	commands := []types.Command{
		&ArduinoPreprocessorRunner{},
	}

	if err := ctx.PreprocPath.MkdirAll(); err != nil {
		return i18n.WrapError(err)
	}

	if ctx.CodeCompleteAt != "" {
		commands = append(commands, &OutputCodeCompletions{})
	} else {
		commands = append(commands, &SketchSaver{})
	}

	GCCPreprocRunner(ctx, sourceFile, ctx.PreprocPath.Join(constants.FILE_CTAGS_TARGET_FOR_GCC_MINUS_E), ctx.IncludeFolders)

	for _, command := range commands {
		PrintRingNameIfDebug(ctx, command)
		err := command.Run(ctx)
		if err != nil {
			return i18n.WrapError(err)
		}
	}

	return nil
}

type ArduinoPreprocessorRunner struct{}

func (s *ArduinoPreprocessorRunner) Run(ctx *types.Context) error {
	buildProperties := ctx.BuildProperties
	targetFilePath := ctx.PreprocPath.Join(constants.FILE_CTAGS_TARGET_FOR_GCC_MINUS_E)
	logger := ctx.GetLogger()

	properties := buildProperties.Clone()
	toolProps := buildProperties.SubTree("tools").SubTree("arduino-preprocessor")
	properties.Merge(toolProps)
	properties.SetPath(constants.BUILD_PROPERTIES_SOURCE_FILE, targetFilePath)
	if ctx.CodeCompleteAt != "" {
		if runtime.GOOS == "windows" {
			//use relative filepath to avoid ":" escaping
			splt := strings.Split(ctx.CodeCompleteAt, ":")
			if len(splt) == 3 {
				//all right, do nothing
			} else {
				splt[1] = filepath.Base(splt[0] + ":" + splt[1])
				ctx.CodeCompleteAt = strings.Join(splt[1:], ":")
			}
		}
		properties.Set("codecomplete", "-output-code-completions="+ctx.CodeCompleteAt)
	} else {
		properties.Set("codecomplete", "")
	}

	pattern := properties.Get(constants.BUILD_PROPERTIES_PATTERN)
	if pattern == constants.EMPTY_STRING {
		return i18n.ErrorfWithLogger(logger, constants.MSG_PATTERN_MISSING, "arduino-preprocessor")
	}

	commandLine := properties.ExpandPropsInString(pattern)
	command, err := utils.PrepareCommand(commandLine, logger, "")
	if err != nil {
		return i18n.WrapError(err)
	}

	if runtime.GOOS == "windows" {
		// chdir in the uppermost directory to avoid UTF-8 bug in clang (https://github.com/arduino/arduino-preprocessor/issues/2)
		command.Dir = filepath.VolumeName(command.Args[0]) + "/"
		//command.Args[0], _ = filepath.Rel(command.Dir, command.Args[0])
	}

	verbose := ctx.Verbose
	if verbose {
		fmt.Println(commandLine)
	}

	buf, err := command.Output()
	if err != nil {
		return errors.New(i18n.WrapError(err).Error() + string(err.(*exec.ExitError).Stderr))
	}

	result := utils.NormalizeUTF8(buf)

	//fmt.Printf("PREPROCESSOR OUTPUT:\n%s\n", output)
	if ctx.CodeCompleteAt != "" {
		ctx.CodeCompletions = string(result)
	} else {
		ctx.Source = string(result)
	}
	return nil
}

type OutputCodeCompletions struct{}

func (s *OutputCodeCompletions) Run(ctx *types.Context) error {
	if ctx.CodeCompletions == "" {
		// we assume it is a json, let's make it compliant at least
		ctx.CodeCompletions = "[]"
	}
	ctx.GetLogger().Println(constants.LOG_LEVEL_INFO, ctx.CodeCompletions)
	return nil
}
