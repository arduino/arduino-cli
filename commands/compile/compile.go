/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package compile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	builder "github.com/arduino/arduino-builder"
	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/configs"
	"github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-map"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// InitCommand prepares the command.
func InitCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "compile",
		Short:   "Compiles Arduino sketches.",
		Long:    "Compiles Arduino sketches.",
		Example: "  " + commands.AppName + " compile -b arduino:avr:uno /home/user/Arduino/MySketch",
		Args:    cobra.MaximumNArgs(1),
		Run:     run,
	}
	command.Flags().StringVarP(&flags.fqbn, "fqbn", "b", "", "Fully Qualified Board Name, e.g.: arduino:avr:uno")
	command.Flags().BoolVar(&flags.showProperties, "show-properties", false, "Show all build properties used instead of compiling.")
	command.Flags().BoolVar(&flags.preprocess, "preprocess", false, "Print preprocessed code to stdout instead of compiling.")
	command.Flags().StringVar(&flags.buildCachePath, "build-cache-path", "", "Builds of 'core.a' are saved into this path to be cached and reused.")
	command.Flags().StringVar(&flags.buildPath, "build-path", "", "Path where to save compiled files. If omitted, a directory will be created in the default temporary path of your OS.")
	command.Flags().StringSliceVar(&flags.buildProperties, "build-properties", []string{}, "List of custom build properties separated by commas. Or can be used multiple times for multiple properties.")
	command.Flags().StringVar(&flags.warnings, "warnings", "none", `Optional, can be "none", "default", "more" and "all". Defaults to "none". Used to tell gcc which warning level to use (-W flag).`)
	command.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Optional, turns on verbose mode.")
	command.Flags().BoolVar(&flags.quiet, "quiet", false, "Optional, supresses almost every output.")
	command.Flags().StringVar(&flags.vidPid, "vid-pid", "", "When specified, VID/PID specific build properties are used, if boards supports them.")
	return command
}

var flags struct {
	fqbn            string   // Fully Qualified Board Name, e.g.: arduino:avr:uno.
	showProperties  bool     // Show all build preferences used instead of compiling.
	preprocess      bool     // Print preprocessed code to stdout.
	buildCachePath  string   // Builds of 'core.a' are saved into this path to be cached and reused.
	buildPath       string   // Path where to save compiled files.
	buildProperties []string // List of custom build properties separated by commas. Or can be used multiple times for multiple properties.
	warnings        string   // Used to tell gcc which warning level to use.
	verbose         bool     // Turns on verbose mode.
	quiet           bool     // Suppresses almost every output.
	vidPid          string   // VID/PID specific build properties.
}

func run(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino compile`")
	var sketchPath *paths.Path
	if len(args) > 0 {
		sketchPath = paths.New(args[0])
	}
	sketch, err := commands.InitSketch(sketchPath)
	if err != nil {
		formatter.PrintError(err, "Error opening sketch.")
		os.Exit(commands.ErrGeneric)
	}

	fqbn := flags.fqbn
	if fqbn == "" && sketch != nil {
		fqbn = sketch.Metadata.CPU.Fqbn
	}
	if fqbn == "" {
		formatter.PrintErrorMessage("No Fully Qualified Board Name provided.")
		os.Exit(commands.ErrGeneric)
	}
	fqbnParts := strings.Split(fqbn, ":")
	if len(fqbnParts) < 3 || len(fqbnParts) > 4 {
		formatter.PrintErrorMessage("Fully Qualified Board Name has incorrect format.")
		os.Exit(commands.ErrBadArgument)
	}
	packageName := fqbnParts[0]
	coreName := fqbnParts[1]

	pm := commands.InitPackageManager()

	// Check for ctags tool
	loadBuiltinCtagsMetadata(pm)
	ctags, _ := getBuiltinCtagsTool(pm)
	if !ctags.IsInstalled() {
		formatter.Print("Downloading and installing missing tool: " + ctags.String())
		core.DownloadToolRelease(pm, ctags)
		core.InstallToolRelease(pm, ctags)

		if err := pm.LoadHardware(commands.Config); err != nil {
			formatter.PrintError(err, "Could not load hardware packages.")
			os.Exit(commands.ErrCoreConfig)
		}
		ctags, _ = getBuiltinCtagsTool(pm)
		if !ctags.IsInstalled() {
			formatter.PrintErrorMessage("Missing ctags tool.")
			os.Exit(commands.ErrCoreConfig)
		}
	}

	targetPlatform := pm.FindPlatform(&packagemanager.PlatformReference{
		Package:              packageName,
		PlatformArchitecture: coreName,
	})
	if targetPlatform == nil || targetPlatform.GetInstalled() == nil {
		formatter.PrintErrorMessage(fmt.Sprintf("\"%[1]s:%[2]s\" platform is not installed, please install it by running \""+commands.AppName+" core install %[1]s:%[2]s\".", packageName, coreName))
		os.Exit(commands.ErrCoreConfig)
	}

	ctx := &types.Context{}

	if parsedFqbn, err := cores.ParseFQBN(fqbn); err != nil {
		formatter.PrintError(err, "Error parsing FQBN.")
		os.Exit(commands.ErrBadArgument)
	} else {
		ctx.FQBN = parsedFqbn
	}
	ctx.SketchLocation = paths.New(sketch.FullPath)

	// FIXME: This will be redundant when arduino-builder will be part of the cli
	if packagesDir, err := commands.Config.HardwareDirectories(); err == nil {
		ctx.HardwareDirs = packagesDir
	} else {
		formatter.PrintError(err, "Cannot get hardware directories.")
		os.Exit(commands.ErrCoreConfig)
	}

	if toolsDir, err := configs.BundleToolsDirectories(); err == nil {
		ctx.ToolsDirs = toolsDir
	} else {
		formatter.PrintError(err, "Cannot get bundled tools directories.")
		os.Exit(commands.ErrCoreConfig)
	}

	ctx.OtherLibrariesDirs = paths.NewPathList()
	ctx.OtherLibrariesDirs.Add(commands.Config.LibrariesDir())

	if flags.buildPath != "" {
		ctx.BuildPath = paths.New(flags.buildPath)
		err = ctx.BuildPath.MkdirAll()
		if err != nil {
			formatter.PrintError(err, "Cannot create the build directory.")
			os.Exit(commands.ErrBadCall)
		}
	}

	ctx.Verbose = flags.verbose

	ctx.CoreBuildCachePath = paths.TempDir().Join("arduino-core-cache")

	ctx.USBVidPid = flags.vidPid
	ctx.WarningsLevel = flags.warnings

	if commands.GlobalFlags.Debug {
		ctx.DebugLevel = 100
	} else {
		ctx.DebugLevel = 5
	}

	ctx.CustomBuildProperties = append(flags.buildProperties, "build.warn_data_percentage=75")

	if flags.buildCachePath != "" {
		ctx.BuildCachePath = paths.New(flags.buildCachePath)
		err = ctx.BuildCachePath.MkdirAll()
		if err != nil {
			formatter.PrintError(err, "Cannot create the build cache directory.")
			os.Exit(commands.ErrBadCall)
		}
	}

	// Will be deprecated.
	ctx.ArduinoAPIVersion = "10607"

	// Check if Arduino IDE is installed and get it's libraries location.
	preferencesTxt := commands.Config.DataDir.Join("preferences.txt")
	ideProperties, err := properties.LoadFromPath(preferencesTxt)
	if err == nil {
		lastIdeSubProperties := ideProperties.SubTree("last").SubTree("ide")
		// Preferences can contain records from previous IDE versions. Find the latest one.
		var pathVariants []string
		for k := range lastIdeSubProperties {
			if strings.HasSuffix(k, ".hardwarepath") {
				pathVariants = append(pathVariants, k)
			}
		}
		sort.Strings(pathVariants)
		ideHardwarePath := lastIdeSubProperties[pathVariants[len(pathVariants)-1]]
		ideLibrariesPath := filepath.Join(filepath.Dir(ideHardwarePath), "libraries")
		ctx.BuiltInLibrariesDirs = paths.NewPathList(ideLibrariesPath)
	}

	if flags.showProperties {
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

	// FIXME: Make a function to obtain these info...
	outputPath := ctx.BuildProperties.ExpandPropsInString("{build.path}/{recipe.output.tmp_file}")
	ext := filepath.Ext(outputPath)
	// FIXME: Make a function to produce a better name...
	fqbn = strings.Replace(fqbn, ":", ".", -1)

	// Copy .hex file to sketch directory
	srcHex := paths.New(outputPath)
	dstHex := paths.New(sketch.FullPath).Join(sketch.Name + "." + fqbn + ext)
	logrus.WithField("from", srcHex).WithField("to", dstHex).Print("copying sketch build output")
	if err = srcHex.CopyTo(dstHex); err != nil {
		formatter.PrintError(err, "Error copying output file.")
		os.Exit(commands.ErrGeneric)
	}

	// Copy .elf file to sketch directory
	srcElf := paths.New(outputPath[:len(outputPath)-3] + "elf")
	dstElf := paths.New(sketch.FullPath).Join(sketch.Name + "." + fqbn + ".elf")
	logrus.WithField("from", srcElf).WithField("to", dstElf).Print("copying sketch build output")
	if err = srcElf.CopyTo(dstElf); err != nil {
		formatter.PrintError(err, "Error copying elf file.")
		os.Exit(commands.ErrGeneric)
	}
}
