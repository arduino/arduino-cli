package compile

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	builder "github.com/arduino/arduino-builder"
	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/cli"
	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/rpc"
	paths "github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	flags "github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

func Compile(ctx context.Context, req *rpc.CompileReq) (*rpc.CompileResp, error) {
	logrus.Info("Executing `arduino compile`")
	var sketchPath *paths.Path
	if len(args) > 0 {
		sketchPath = paths.New(args[0])
	}
	sketch, err := cli.InitSketch(sketchPath)
	if err != nil {
		formatter.PrintError(err, "Error opening sketch.")
		os.Exit(cli.ErrGeneric)
	}

	if flags.fqbn == "" && sketch != nil && sketch.Metadata != nil {
		flags.fqbn = sketch.Metadata.CPU.Fqbn
	}
	if flags.fqbn == "" {
		formatter.PrintErrorMessage("No Fully Qualified Board Name provided.")
		os.Exit(cli.ErrGeneric)
	}
	fqbn, err := cores.ParseFQBN(flags.fqbn)
	if err != nil {
		formatter.PrintErrorMessage("Fully Qualified Board Name has incorrect format.")
		os.Exit(cli.ErrBadArgument)
	}

	pm, _ := cli.InitPackageAndLibraryManager()

	// Check for ctags tool
	loadBuiltinCtagsMetadata(pm)
	ctags, _ := getBuiltinCtagsTool(pm)
	if !ctags.IsInstalled() {
		formatter.Print("Downloading and installing missing tool: " + ctags.String())
		core.DownloadToolRelease(pm, ctags)
		core.InstallToolRelease(pm, ctags)

		if err := pm.LoadHardware(cli.Config); err != nil {
			formatter.PrintError(err, "Could not load hardware packages.")
			os.Exit(cli.ErrCoreConfig)
		}
		ctags, _ = getBuiltinCtagsTool(pm)
		if !ctags.IsInstalled() {
			formatter.PrintErrorMessage("Missing ctags tool.")
			os.Exit(cli.ErrCoreConfig)
		}
	}

	targetPlatform := pm.FindPlatform(&packagemanager.PlatformReference{
		Package:              fqbn.Package,
		PlatformArchitecture: fqbn.PlatformArch,
	})
	if targetPlatform == nil || pm.GetInstalledPlatformRelease(targetPlatform) == nil {
		errorMessage := fmt.Sprintf(
			"\"%[1]s:%[2]s\" platform is not installed, please install it by running \""+
				cli.AppName+" core install %[1]s:%[2]s\".", fqbn.Package, fqbn.PlatformArch)
		formatter.PrintErrorMessage(errorMessage)
		os.Exit(cli.ErrCoreConfig)
	}

	ctx := &types.Context{}
	ctx.PackageManager = pm
	ctx.FQBN = fqbn
	ctx.SketchLocation = sketch.FullPath

	// FIXME: This will be redundant when arduino-builder will be part of the cli
	if packagesDir, err := cli.Config.HardwareDirectories(); err == nil {
		ctx.HardwareDirs = packagesDir
	} else {
		formatter.PrintError(err, "Cannot get hardware directories.")
		os.Exit(cli.ErrCoreConfig)
	}

	if toolsDir, err := cli.Config.BundleToolsDirectories(); err == nil {
		ctx.ToolsDirs = toolsDir
	} else {
		formatter.PrintError(err, "Cannot get bundled tools directories.")
		os.Exit(cli.ErrCoreConfig)
	}

	ctx.OtherLibrariesDirs = paths.NewPathList()
	ctx.OtherLibrariesDirs.Add(cli.Config.LibrariesDir())

	if flags.buildPath != "" {
		ctx.BuildPath = paths.New(flags.buildPath)
		err = ctx.BuildPath.MkdirAll()
		if err != nil {
			formatter.PrintError(err, "Cannot create the build directory.")
			os.Exit(cli.ErrBadCall)
		}
	}

	ctx.Verbose = flags.verbose

	ctx.CoreBuildCachePath = paths.TempDir().Join("arduino-core-cache")

	ctx.USBVidPid = flags.vidPid
	ctx.WarningsLevel = flags.warnings

	if cli.GlobalFlags.Debug {
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
			os.Exit(cli.ErrBadCall)
		}
	}

	// Will be deprecated.
	ctx.ArduinoAPIVersion = "10607"

	// Check if Arduino IDE is installed and get it's libraries location.
	preferencesTxt := cli.Config.DataDir.Join("preferences.txt")
	ideProperties, err := properties.LoadFromPath(preferencesTxt)
	if err == nil {
		lastIdeSubProperties := ideProperties.SubTree("last").SubTree("ide")
		// Preferences can contain records from previous IDE versions. Find the latest one.
		var pathVariants []string
		for k := range lastIdeSubProperties.AsMap() {
			if strings.HasSuffix(k, ".hardwarepath") {
				pathVariants = append(pathVariants, k)
			}
		}
		sort.Strings(pathVariants)
		ideHardwarePath := lastIdeSubProperties.Get(pathVariants[len(pathVariants)-1])
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
		os.Exit(cli.ErrGeneric)
	}

	// FIXME: Make a function to obtain these info...
	outputPath := ctx.BuildProperties.ExpandPropsInString("{build.path}/{recipe.output.tmp_file}")
	ext := filepath.Ext(outputPath)

	// FIXME: Make a function to produce a better name...
	// Make the filename without the FQBN configs part
	fqbn.Configs = properties.NewMap()
	fqbnSuffix := strings.Replace(fqbn.String(), ":", ".", -1)

	var exportPath *paths.Path
	var exportFile string
	if flags.exportFile == "" {
		exportPath = sketch.FullPath
		exportFile = sketch.Name + "." + fqbnSuffix
	} else {
		exportPath = paths.New(flags.exportFile).Parent()
		exportFile = paths.New(flags.exportFile).Base()
		if strings.HasSuffix(exportFile, ext) {
			exportFile = exportFile[:len(exportFile)-len(ext)]
		}
	}

	// Copy .hex file to sketch directory
	srcHex := paths.New(outputPath)
	dstHex := exportPath.Join(exportFile + ext)
	logrus.WithField("from", srcHex).WithField("to", dstHex).Print("copying sketch build output")
	if err = srcHex.CopyTo(dstHex); err != nil {
		formatter.PrintError(err, "Error copying output file.")
		os.Exit(cli.ErrGeneric)
	}

	// Copy .elf file to sketch directory
	srcElf := paths.New(outputPath[:len(outputPath)-3] + "elf")
	dstElf := exportPath.Join(exportFile + ".elf")
	logrus.WithField("from", srcElf).WithField("to", dstElf).Print("copying sketch build output")
	if err = srcElf.CopyTo(dstElf); err != nil {
		formatter.PrintError(err, "Error copying elf file.")
		os.Exit(cli.ErrGeneric)
	}
}
