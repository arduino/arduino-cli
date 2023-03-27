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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	bldr "github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

// ArduinoPreprocessorProperties are the platform properties needed to run arduino-preprocessor
var ArduinoPreprocessorProperties = properties.NewFromHashmap(map[string]string{
	// Ctags
	"tools.arduino-preprocessor.path":     "{runtime.tools.arduino-preprocessor.path}",
	"tools.arduino-preprocessor.cmd.path": "{path}/arduino-preprocessor",
	"tools.arduino-preprocessor.pattern":  `"{cmd.path}" "{source_file}" -- -std=gnu++11`,

	"preproc.macros.flags": "-w -x c++ -E -CC",
})

type PreprocessSketchArduino struct{}

func (s *PreprocessSketchArduino) Run(ctx *types.Context) error {
	sourceFile := ctx.SketchBuildPath.Join(ctx.Sketch.MainFile.Base() + ".cpp")
	commands := []types.Command{
		&ArduinoPreprocessorRunner{},
	}

	if err := ctx.PreprocPath.MkdirAll(); err != nil {
		return errors.WithStack(err)
	}

	GCCPreprocRunner(ctx, sourceFile, ctx.PreprocPath.Join(constants.FILE_CTAGS_TARGET_FOR_GCC_MINUS_E), ctx.IncludeFolders)

	for _, command := range commands {
		PrintRingNameIfDebug(ctx, command)
		err := command.Run(ctx)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return bldr.SketchSaveItemCpp(ctx.Sketch.MainFile, []byte(ctx.Source), ctx.SketchBuildPath)
}

type ArduinoPreprocessorRunner struct{}

func (s *ArduinoPreprocessorRunner) Run(ctx *types.Context) error {
	buildProperties := ctx.BuildProperties
	targetFilePath := ctx.PreprocPath.Join(constants.FILE_CTAGS_TARGET_FOR_GCC_MINUS_E)

	preprocProperties := buildProperties.Clone()
	toolProps := buildProperties.SubTree("tools").SubTree("arduino-preprocessor")
	preprocProperties.Merge(toolProps)
	preprocProperties.SetPath(constants.BUILD_PROPERTIES_SOURCE_FILE, targetFilePath)

	pattern := preprocProperties.Get(constants.BUILD_PROPERTIES_PATTERN)
	if pattern == constants.EMPTY_STRING {
		return errors.New(tr("arduino-preprocessor pattern is missing"))
	}

	commandLine := preprocProperties.ExpandPropsInString(pattern)
	parts, err := properties.SplitQuotedString(commandLine, `"'`, false)
	if err != nil {
		return errors.WithStack(err)
	}
	command := exec.Command(parts[0], parts[1:]...)
	command.Env = append(os.Environ(), ctx.PackageManager.GetEnvVarsForSpawnedProcess()...)

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
		return errors.New(errors.WithStack(err).Error() + string(err.(*exec.ExitError).Stderr))
	}

	result := utils.NormalizeUTF8(buf)

	//fmt.Printf("PREPROCESSOR OUTPUT:\n%s\n", output)
	ctx.Source = string(result)
	return nil
}
