package types

import (
	"strings"

	"github.com/arduino/arduino-builder/i18n"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesresolver"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-map"
)

type ProgressStruct struct {
	PrintEnabled bool
	Steps        float64
	Progress     float64
}

// Context structure
type Context struct {
	// Build options
	HardwareDirs         paths.PathList
	ToolsDirs            paths.PathList
	BuiltInToolsDirs     paths.PathList
	BuiltInLibrariesDirs paths.PathList
	OtherLibrariesDirs   paths.PathList
	SketchLocation       *paths.Path
	WatchedLocations     paths.PathList
	ArduinoAPIVersion    string
	FQBN                 *cores.FQBN
	CodeCompleteAt       string

	// Build options are serialized here
	BuildOptionsJson         string
	BuildOptionsJsonPrevious string

	PackageManager *packagemanager.PackageManager
	Hardware       *cores.Packages
	AllTools       []*cores.ToolRelease
	RequiredTools  []*cores.ToolRelease
	TargetBoard    *cores.Board
	TargetPackage  *cores.Package
	TargetPlatform *cores.PlatformRelease
	ActualPlatform *cores.PlatformRelease
	USBVidPid      string

	PlatformKeyRewrites    PlatforKeysRewrite
	HardwareRewriteResults map[*cores.PlatformRelease][]PlatforKeyRewrite

	BuildProperties      properties.Map
	BuildCore            string
	BuildPath            *paths.Path
	BuildCachePath       *paths.Path
	SketchBuildPath      *paths.Path
	CoreBuildPath        *paths.Path
	CoreBuildCachePath   *paths.Path
	CoreArchiveFilePath  *paths.Path
	CoreObjectsFiles     paths.PathList
	LibrariesBuildPath   *paths.Path
	LibrariesObjectFiles paths.PathList
	PreprocPath          *paths.Path
	SketchObjectFiles    paths.PathList

	CollectedSourceFiles *UniqueSourceFileQueue

	Sketch          *Sketch
	Source          string
	SourceGccMinusE string
	CodeCompletions string

	WarningsLevel string

	// Libraries handling
	LibrariesManager           *librariesmanager.LibrariesManager
	LibrariesResolver          *librariesresolver.Cpp
	ImportedLibraries          libraries.List
	LibrariesResolutionResults map[string]LibraryResolutionResult
	IncludeFolders             paths.PathList
	//OutputGccMinusM            string

	// C++ Parsing
	CTagsOutput                 string
	CTagsTargetFile             *paths.Path
	CTagsOfPreprocessedSource   []*CTag
	IncludeSection              string
	LineOffset                  int
	PrototypesSection           string
	PrototypesLineWhereToInsert int
	Prototypes                  []*Prototype

	// Verbosity settings
	Verbose           bool
	DebugPreprocessor bool

	// Dry run, only create progress map
	Progress ProgressStruct

	// Contents of a custom build properties file (line by line)
	CustomBuildProperties []string

	// Logging
	logger     i18n.Logger
	DebugLevel int

	// Reuse old tools since the backing storage didn't change
	CanUseCachedTools bool

	// Experimental: use arduino-preprocessor to create prototypes
	UseArduinoPreprocessor bool
}

func (ctx *Context) ExtractBuildOptions() properties.Map {
	opts := make(properties.Map)
	opts["hardwareFolders"] = strings.Join(ctx.HardwareDirs.AsStrings(), ",")
	opts["toolsFolders"] = strings.Join(ctx.ToolsDirs.AsStrings(), ",")
	opts["builtInLibrariesFolders"] = strings.Join(ctx.BuiltInLibrariesDirs.AsStrings(), ",")
	opts["otherLibrariesFolders"] = strings.Join(ctx.OtherLibrariesDirs.AsStrings(), ",")
	opts["sketchLocation"] = ctx.SketchLocation.String()
	var additionalFilesRelative []string
	if ctx.Sketch != nil {
		for _, sketch := range ctx.Sketch.AdditionalFiles {
			absPath := ctx.SketchLocation.Parent()
			relPath, err := sketch.Name.RelTo(absPath)
			if err != nil {
				continue // ignore
			}
			additionalFilesRelative = append(additionalFilesRelative, relPath.String())
		}
	}
	opts["fqbn"] = ctx.FQBN.String()
	opts["runtime.ide.version"] = ctx.ArduinoAPIVersion
	opts["customBuildProperties"] = strings.Join(ctx.CustomBuildProperties, ",")
	opts["additionalFiles"] = strings.Join(additionalFilesRelative, ",")
	return opts
}

func (ctx *Context) InjectBuildOptions(opts properties.Map) {
	ctx.HardwareDirs = paths.NewPathList(strings.Split(opts["hardwareFolders"], ",")...)
	ctx.ToolsDirs = paths.NewPathList(strings.Split(opts["toolsFolders"], ",")...)
	ctx.BuiltInLibrariesDirs = paths.NewPathList(strings.Split(opts["builtInLibrariesFolders"], ",")...)
	ctx.OtherLibrariesDirs = paths.NewPathList(strings.Split(opts["otherLibrariesFolders"], ",")...)
	ctx.SketchLocation = paths.New(opts["sketchLocation"])
	fqbn, err := cores.ParseFQBN(opts["fqbn"])
	if err != nil {
		i18n.ErrorfWithLogger(ctx.GetLogger(), "Error in FQBN: %s", err)
	}
	ctx.FQBN = fqbn
	ctx.ArduinoAPIVersion = opts["runtime.ide.version"]
	ctx.CustomBuildProperties = strings.Split(opts["customBuildProperties"], ",")
}

func (ctx *Context) GetLogger() i18n.Logger {
	if ctx.logger == nil {
		return &i18n.HumanLogger{}
	}
	return ctx.logger
}

func (ctx *Context) SetLogger(l i18n.Logger) {
	ctx.logger = l
}
