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

	bldr "github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/pkg/errors"
)

func PreprocessSketchWithCtags(ctx *types.Context) error {
	// Generate the full pathname for the preproc output file
	if err := ctx.PreprocPath.MkdirAll(); err != nil {
		return errors.WithStack(err)
	}
	targetFilePath := ctx.PreprocPath.Join("sketch_merged.cpp")

	// Run preprocessor
	sourceFile := ctx.SketchBuildPath.Join(ctx.Sketch.MainFile.Base() + ".cpp")
	if err := GCCPreprocRunner(ctx, sourceFile, targetFilePath, ctx.IncludeFolders); err != nil {
		if !ctx.OnlyUpdateCompilationDatabase {
			return errors.WithStack(err)
		}

		// Do not bail out if we are generating the compile commands database
		ctx.Info(
			fmt.Sprintf("%s: %s",
				tr("An error occurred adding prototypes"),
				tr("the compilation database may be incomplete or inaccurate")))
		if err := sourceFile.CopyTo(targetFilePath); err != nil {
			return errors.WithStack(err)
		}
	}

	if src, err := targetFilePath.ReadFile(); err != nil {
		return err
	} else {
		ctx.SketchSourceAfterCppPreprocessing = string(src)
	}

	commands := []types.Command{
		&FilterSketchSource{Source: &ctx.SketchSourceAfterCppPreprocessing},
		&CTagsRunner{Source: &ctx.SketchSourceAfterCppPreprocessing, TargetFileName: "sketch_merged.cpp"},
		&PrototypesAdder{},
	}

	for _, command := range commands {
		PrintRingNameIfDebug(ctx, command)
		err := command.Run(ctx)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	if err := bldr.SketchSaveItemCpp(ctx.Sketch.MainFile, []byte(ctx.SketchSourceAfterArduinoPreprocessing), ctx.SketchBuildPath); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
