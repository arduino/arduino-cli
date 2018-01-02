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

package compile

import (
	"os"
	"path/filepath"
	"strings"

	builder "github.com/arduino/arduino-builder"
	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/arduino-builder/utils"
	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Init prepares the command.
func Init(rootCommand *cobra.Command) {
	rootCommand.AddCommand(command)
	command.Flags().StringVarP(&flags.fullyQualifiedBoardName, "fqbn", "b", "", "Fully Qualified Board Name, e.g.: arduino:avr:uno")
	command.Flags().BoolVar(&flags.dumpPreferences, "dump-prefs", false, "Show all build preferences used instead of compiling.")
	command.Flags().BoolVar(&flags.preprocess, "preprocess", false, "Print preprocessed code to stdout instead of compiling.")
	command.Flags().StringVar(&flags.buildPath, "build-path", "", "Folder where to save compiled files. If omitted, a folder will be created in the temporary folder specified by your OS.")
	command.Flags().StringVar(&flags.warnings, "warnings", "none", `Optional, can be "none", "default", "more" and "all". Defaults to "none". Used to tell gcc which warning level to use (-W flag).`)
	command.Flags().BoolVar(&flags.verbose, "verbose", false, "Optional, turns on verbose mode.")
	command.Flags().BoolVar(&flags.quiet, "quiet", false, "Optional, supresses almost every output.")
	command.Flags().IntVar(&flags.debugLevel, "debug-level", 5, "Optional, defaults to 5. Used for debugging. Set it to 10 when submitting an issue.")
	command.Flags().StringVar(&flags.vidPid, "vid-pid", "", "When specified, VID/PID specific build properties are used, if boards supports them.")
}

var flags struct {
	fullyQualifiedBoardName string // Fully Qualified Board Name, e.g.: arduino:avr:uno.
	dumpPreferences         bool   // Show all build preferences used instead of compiling.
	preprocess              bool   // Print preprocessed code to stdout.
	buildPath               string // Folder where to save compiled files.
	warnings                string // Used to tell gcc which warning level to use.
	verbose                 bool   // Turns on verbose mode.
	quiet                   bool   // Supresses almost every output.
	debugLevel              int    // Used for debugging.
	vidPid                  string // VID/PID specific build properties.
}

var command = &cobra.Command{
	Use:     "compile",
	Short:   "Compiles Arduino sketches.",
	Long:    "Compiles Arduino sketches. The specified sketch must be downloaded prior to compile.",
	Example: "arduino compile sketchName",
	Args:    cobra.ExactArgs(1),
	Run:     run,
}

// TODO: check if ctags or core is missing.
func run(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino compile`")
	isCorrectSyntax := true
	// TODO: Check if sketch exists.
	sketchName := args[0]
	var packageName string
	if flags.fullyQualifiedBoardName == "" {
		// TODO: Determine from Sketch file.
		formatter.PrintErrorMessage("No Fully Qualified Board Name provided.")
		isCorrectSyntax = false
	} else {
		fqbnParts := strings.Split(flags.fullyQualifiedBoardName, ":")
		if len(fqbnParts) != 3 {
			formatter.PrintErrorMessage("Fully Qualified Board Name has incorrect format.")
			isCorrectSyntax = false
		} else {
			packageName = fqbnParts[0]
		}
	}
	if !isCorrectSyntax {
		os.Exit(commands.ErrBadCall)
	}

	ctx := &types.Context{}

	ctx.FQBN = flags.fullyQualifiedBoardName
	ctx.SketchLocation = filepath.Join(common.SketchbookFolder, sketchName)

	packagesFolder, err := common.GetDefaultPkgFolder()
	if err != nil {
		formatter.PrintError(err, "Cannot get packages folder.")
		os.Exit(commands.ErrCoreConfig)
	}
	ctx.HardwareFolders = []string{packagesFolder}

	toolsFolder, err := common.GetDefaultToolsFolder(packageName)
	if err != nil {
		formatter.PrintError(err, "Cannot get tools folder.")
		os.Exit(commands.ErrCoreConfig)
	}
	ctx.ToolsFolders = []string{toolsFolder}

	librariesFolder, err := common.GetDefaultLibFolder()
	if err != nil {
		formatter.PrintError(err, "Cannot get libraries folder.")
		os.Exit(commands.ErrCoreConfig)
	}
	ctx.OtherLibrariesFolders = []string{librariesFolder}

	ctx.BuildPath = flags.buildPath
	if ctx.BuildPath != "" {
		// TODO: Needed?
		/*_, err = os.Stat(ctx.BuildPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(commands.ErrBadCall)
		}*/

		err = utils.EnsureFolderExists(ctx.BuildPath)
		if err != nil {
			formatter.PrintError(err, "Cannot create the build folder.")
			os.Exit(commands.ErrBadCall)
		}
	}

	ctx.Verbose = flags.verbose
	ctx.DebugLevel = flags.debugLevel

	ctx.USBVidPid = flags.vidPid
	ctx.WarningsLevel = flags.warnings

	// TODO:
	ctx.ArduinoAPIVersion = "10600"
	//ctx.BuiltInLibrariesFolders = ?
	//ctx.CustomBuildProperties = ?
	//ctx.BuildCachePath = ?

	if flags.dumpPreferences {
		err = builder.RunParseHardwareAndDumpBuildProperties(ctx)
	} else if flags.preprocess {
		err = builder.RunPreprocess(ctx)
	} else {
		err = builder.RunBuilder(ctx)
	}

	if err != nil {
		formatter.PrintError(err, "Compilation failed.")
		os.Exit(commands.ErrGeneric)
	}
}
