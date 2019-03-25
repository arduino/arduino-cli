package compile

import (
	"context"
	"fmt"
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
	"github.com/sirupsen/logrus"
)

func Compile(ctx context.Context, req *rpc.CompileReq) (*rpc.CompileResp, error) {
	logrus.Info("Executing `arduino compile`")
	var sketchPath *paths.Path
	if req.GetSketchPath() != "" {
		sketchPath = paths.New(req.GetSketchPath())
	}
	sketch, err := cli.InitSketch(sketchPath)
	if err != nil {
		return &rpc.CompileResp{
			Result: rpc.Error("Error opening sketch", rpc.ErrGeneric),
		}, nil
	}

	fqbnIn := req.GetFqbn()
	if fqbnIn == "" && sketch != nil && sketch.Metadata != nil {
		fqbnIn = sketch.Metadata.CPU.Fqbn
	}
	if fqbnIn == "" {
		formatter.PrintErrorMessage("No Fully Qualified Board Name provided.")
		return &rpc.CompileResp{
			Result: rpc.Error("No Fully Qualified Board Name provided.", rpc.ErrGeneric),
		}, nil
	}
	fqbn, err := cores.ParseFQBN(fqbnIn)
	if err != nil {
		formatter.PrintErrorMessage("Fully Qualified Board Name has incorrect format.")
		return &rpc.CompileResp{
			Result: rpc.Error("Fully Qualified Board Name has incorrect format.", rpc.ErrGeneric),
		}, nil
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
			return &rpc.CompileResp{
				Result: rpc.Error("Could not load hardware packages.", rpc.ErrGeneric),
			}, nil
		}
		ctags, _ = getBuiltinCtagsTool(pm)
		if !ctags.IsInstalled() {
			formatter.PrintErrorMessage("Missing ctags tool.")
			return &rpc.CompileResp{
				Result: rpc.Error("Missing ctags tool.", rpc.ErrGeneric),
			}, nil
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
		return &rpc.CompileResp{
			Result: rpc.Error(errorMessage, rpc.ErrGeneric),
		}, nil
	}

	builderCtx := &types.Context{}
	builderCtx.PackageManager = pm
	builderCtx.FQBN = fqbn
	builderCtx.SketchLocation = sketch.FullPath

	// FIXME: This will be redundant when arduino-builder will be part of the cli
	if packagesDir, err := cli.Config.HardwareDirectories(); err == nil {
		builderCtx.HardwareDirs = packagesDir
	} else {
		return &rpc.CompileResp{
			Result: rpc.Error("Cannot get hardware directories.", rpc.ErrGeneric),
		}, nil
	}

	if toolsDir, err := cli.Config.BundleToolsDirectories(); err == nil {
		builderCtx.ToolsDirs = toolsDir
	} else {
		return &rpc.CompileResp{
			Result: rpc.Error("Cannot get bundled tools directories.", rpc.ErrGeneric),
		}, nil
	}

	builderCtx.OtherLibrariesDirs = paths.NewPathList()
	builderCtx.OtherLibrariesDirs.Add(cli.Config.LibrariesDir())

	if req.GetBuildPath() != "" {
		builderCtx.BuildPath = paths.New(req.GetBuildPath())
		err = builderCtx.BuildPath.MkdirAll()
		if err != nil {
			return &rpc.CompileResp{
				Result: rpc.Error("Cannot create the build directory.", rpc.ErrGeneric),
			}, nil
		}
	}

	builderCtx.Verbose = req.GetVerbose()

	builderCtx.CoreBuildCachePath = paths.TempDir().Join("arduino-core-cache")

	builderCtx.USBVidPid = req.GetVidPid()
	builderCtx.WarningsLevel = req.GetWarnings()

	if cli.GlobalFlags.Debug {
		builderCtx.DebugLevel = 100
	} else {
		builderCtx.DebugLevel = 5
	}

	builderCtx.CustomBuildProperties = append(req.GetBuildProperties(), "build.warn_data_percentage=75")

	if req.GetBuildCachePath() != "" {
		builderCtx.BuildCachePath = paths.New(req.GetBuildCachePath())
		err = builderCtx.BuildCachePath.MkdirAll()
		if err != nil {
			return &rpc.CompileResp{
				Result: rpc.Error("Cannot create the build cache directory.", rpc.ErrGeneric),
			}, nil
		}
	}

	// Will be deprecated.
	builderCtx.ArduinoAPIVersion = "10607"

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
		builderCtx.BuiltInLibrariesDirs = paths.NewPathList(ideLibrariesPath)
	}

	if req.GetShowProperties() {
		err = builder.RunParseHardwareAndDumpBuildProperties(builderCtx)
	} else if req.GetPreprocess() {
		err = builder.RunPreprocess(builderCtx)
	} else {
		err = builder.RunBuilder(builderCtx)
	}

	if err != nil {
		return &rpc.CompileResp{
			Result: rpc.Error("Compilation failed.", rpc.ErrGeneric),
		}, nil
	}

	// FIXME: Make a function to obtain these info...
	outputPath := builderCtx.BuildProperties.ExpandPropsInString("{build.path}/{recipe.output.tmp_file}")
	ext := filepath.Ext(outputPath)

	// FIXME: Make a function to produce a better name...
	// Make the filename without the FQBN configs part
	fqbn.Configs = properties.NewMap()
	fqbnSuffix := strings.Replace(fqbn.String(), ":", ".", -1)

	var exportPath *paths.Path
	var exportFile string
	if req.GetExportFile() == "" {
		exportPath = sketch.FullPath
		exportFile = sketch.Name + "." + fqbnSuffix
	} else {
		exportPath = paths.New(req.GetExportFile()).Parent()
		exportFile = paths.New(req.GetExportFile()).Base()
		if strings.HasSuffix(exportFile, ext) {
			exportFile = exportFile[:len(exportFile)-len(ext)]
		}
	}

	// Copy .hex file to sketch directory
	srcHex := paths.New(outputPath)
	dstHex := exportPath.Join(exportFile + ext)
	logrus.WithField("from", srcHex).WithField("to", dstHex).Print("copying sketch build output")
	if err = srcHex.CopyTo(dstHex); err != nil {
		return &rpc.CompileResp{
			Result: rpc.Error("Error copying output file.", rpc.ErrGeneric),
		}, nil
	}

	// Copy .elf file to sketch directory
	srcElf := paths.New(outputPath[:len(outputPath)-3] + "elf")
	dstElf := exportPath.Join(exportFile + ".elf")
	logrus.WithField("from", srcElf).WithField("to", dstElf).Print("copying sketch build output")
	if err = srcElf.CopyTo(dstElf); err != nil {
		formatter.PrintError(err, "Error copying elf file.")
		return &rpc.CompileResp{
			Result: rpc.Error("Error copying elf file.", rpc.ErrGeneric),
		}, nil
	}

	return &rpc.CompileResp{}, nil
}
