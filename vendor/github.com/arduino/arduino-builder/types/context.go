package types

import (
	"strings"

	"github.com/bcmi-labs/arduino-cli/cores/packagemanager"

	"github.com/arduino/arduino-builder/i18n"
	"github.com/arduino/go-properties-map"
	"github.com/bcmi-labs/arduino-cli/cores"
)

// Context structure
type Context struct {
	// Build options
	HardwareFolders         []string
	ToolsFolders            []string
	BuiltInToolsFolders     []string
	LibrariesFolders        []string
	BuiltInLibrariesFolders []string
	OtherLibrariesFolders   []string
	SketchLocation          string
	ArduinoAPIVersion       string
	FQBN                    string

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
	BuildPath            string
	BuildCachePath       string
	SketchBuildPath      string
	CoreBuildPath        string
	CoreBuildCachePath   string
	CoreArchiveFilePath  string
	CoreObjectsFiles     []string
	LibrariesBuildPath   string
	LibrariesObjectFiles []string
	PreprocPath          string
	SketchObjectFiles    []string

	CollectedSourceFiles *UniqueSourceFileQueue

	Sketch          *Sketch
	Source          string
	SourceGccMinusE string

	WarningsLevel string

	// Libraries handling
	Libraries                  []*Library
	HeaderToLibraries          map[string][]*Library
	ImportedLibraries          []*Library
	LibrariesResolutionResults map[string]LibraryResolutionResult
	IncludeJustFound           string
	IncludeFolders             []string

	// C++ Parsing
	CTagsOutput                 string
	CTagsTargetFile             string
	CTagsOfPreprocessedSource   []*CTag
	IncludeSection              string
	LineOffset                  int
	PrototypesSection           string
	PrototypesLineWhereToInsert int
	Prototypes                  []*Prototype

	// Verbosity settings
	Verbose           bool
	DebugPreprocessor bool

	// Contents of a custom build properties file (line by line)
	CustomBuildProperties []string

	// Logging
	logger     i18n.Logger
	DebugLevel int

	// ReadFileAndStoreInContext command
	FileToRead string
}

func (ctx *Context) ExtractBuildOptions() properties.Map {
	opts := make(properties.Map)
	opts["hardwareFolders"] = strings.Join(ctx.HardwareFolders, ",")
	opts["toolsFolders"] = strings.Join(ctx.ToolsFolders, ",")
	opts["builtInLibrariesFolders"] = strings.Join(ctx.BuiltInLibrariesFolders, ",")
	opts["otherLibrariesFolders"] = strings.Join(ctx.OtherLibrariesFolders, ",")
	opts["sketchLocation"] = ctx.SketchLocation
	opts["fqbn"] = ctx.FQBN
	opts["runtime.ide.version"] = ctx.ArduinoAPIVersion
	opts["customBuildProperties"] = strings.Join(ctx.CustomBuildProperties, ",")
	return opts
}

func (ctx *Context) InjectBuildOptions(opts properties.Map) {
	ctx.HardwareFolders = strings.Split(opts["hardwareFolders"], ",")
	ctx.ToolsFolders = strings.Split(opts["toolsFolders"], ",")
	ctx.BuiltInLibrariesFolders = strings.Split(opts["builtInLibrariesFolders"], ",")
	ctx.OtherLibrariesFolders = strings.Split(opts["otherLibrariesFolders"], ",")
	ctx.SketchLocation = opts["sketchLocation"]
	ctx.FQBN = opts["fqbn"]
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
