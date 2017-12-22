/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
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
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package cmd

import (
	"os"
	"path/filepath"
	"strings"

	builder "github.com/arduino/arduino-builder"
	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/arduino-builder/utils"
	"github.com/bcmi-labs/arduino-cli/cmd/formatter"
	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var arduinoCompileCmd = &cobra.Command{
	Use:     `compile`,
	Short:   `Compiles Arduino sketches`,
	Long:    `Compiles Arduino sketches`,
	Example: `arduino compile --fqbn arduino:avr:uno --sketch mySketch`,
	Args:    cobra.NoArgs,
	Run:     executeCompileCommand,
}

// TODO: convert required flags to arguments, install ctags and core if missing
func executeCompileCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino compile`")
	isCorrectSyntax := true
	if arduinoCompileFlags.SketchName == "" {
		formatter.PrintErrorMessage("No sketch name provided")
		isCorrectSyntax = false
	}
	var packageName string
	if arduinoCompileFlags.FullyQualifiedBoardName == "" {
		formatter.PrintErrorMessage("No Fully Qualified Board Name provided")
		isCorrectSyntax = false
	} else {
		fqbnParts := strings.Split(arduinoCompileFlags.FullyQualifiedBoardName, ":")
		if len(fqbnParts) != 3 {
			formatter.PrintErrorMessage("Fully Qualified Board Name has incorrect format")
			isCorrectSyntax = false
		} else {
			packageName = fqbnParts[0]
		}
	}
	if !isCorrectSyntax {
		os.Exit(errBadCall)
	}

	ctx := &types.Context{}

	ctx.FQBN = arduinoCompileFlags.FullyQualifiedBoardName
	ctx.SketchLocation = filepath.Join(common.SketchbookFolder, arduinoCompileFlags.SketchName)

	packagesFolder, err := common.GetDefaultPkgFolder()
	if err != nil {
		formatter.PrintError(err, "Cannot get packages folder")
		os.Exit(errCoreConfig)
	}
	ctx.HardwareFolders = []string{packagesFolder}

	toolsFolder, err := common.GetDefaultToolsFolder(packageName)
	if err != nil {
		formatter.PrintError(err, "Cannot get tools folder")
		os.Exit(errCoreConfig)
	}
	ctx.ToolsFolders = []string{toolsFolder}

	librariesFolder, err := common.GetDefaultLibFolder()
	if err != nil {
		formatter.PrintError(err, "Cannot get libraries folder")
		os.Exit(errCoreConfig)
	}
	ctx.OtherLibrariesFolders = []string{librariesFolder}

	ctx.BuildPath = arduinoCompileFlags.BuildPath
	if ctx.BuildPath != "" {
		// TODO: Needed?
		/*_, err = os.Stat(ctx.BuildPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(errBadCall)
		}*/

		err = utils.EnsureFolderExists(ctx.BuildPath)
		if err != nil {
			formatter.PrintError(err, "Cannot create the build folder")
			os.Exit(errBadCall)
		}
	}

	ctx.Verbose = arduinoCompileFlags.Verbose
	ctx.DebugLevel = arduinoCompileFlags.DebugLevel

	ctx.USBVidPid = arduinoCompileFlags.VidPid
	ctx.WarningsLevel = arduinoCompileFlags.Warnings

	// TODO:
	ctx.ArduinoAPIVersion = "10600"
	//ctx.BuiltInLibrariesFolders = ?
	//ctx.CustomBuildProperties = ?
	//ctx.BuildCachePath = ?

	if arduinoCompileFlags.DumpPreferences {
		err = builder.RunParseHardwareAndDumpBuildProperties(ctx)
	} else if arduinoCompileFlags.Preprocess {
		err = builder.RunPreprocess(ctx)
	} else {
		err = builder.RunBuilder(ctx)
	}

	if err != nil {
		formatter.PrintError(err, "Compilation failed")
		os.Exit(errGeneric)
	}
}
