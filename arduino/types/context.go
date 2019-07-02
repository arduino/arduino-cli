// This file is part of arduino-cli.
//
// Copyright 2019 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.

package types

import (
	"io"
	"strings"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesresolver"
	"github.com/arduino/arduino-cli/legacy/builder/i18n"
	paths "github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
)

// ProgressStruct represents a progress state
type ProgressStruct struct {
	PrintEnabled bool
	Steps        float64
	Progress     float64
}

// LibraryResolutionResult contains the libraries resolution
type LibraryResolutionResult struct {
	Library          *libraries.Library
	NotUsedLibraries []*libraries.Library
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
	BuildOptionsJSON         string
	BuildOptionsJSONPrevious string

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

	BuildProperties      *properties.Map
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

	// Out and Err stream to redirect all Exec commands
	ExecStdout io.Writer
	ExecStderr io.Writer
}

// ExtractBuildOptions returns a properties map containing build options
func (ctx *Context) ExtractBuildOptions() *properties.Map {
	opts := properties.NewMap()
	opts.Set("hardwareFolders", strings.Join(ctx.HardwareDirs.AsStrings(), ","))
	opts.Set("toolsFolders", strings.Join(ctx.ToolsDirs.AsStrings(), ","))
	opts.Set("builtInLibrariesFolders", strings.Join(ctx.BuiltInLibrariesDirs.AsStrings(), ","))
	opts.Set("otherLibrariesFolders", strings.Join(ctx.OtherLibrariesDirs.AsStrings(), ","))
	opts.SetPath("sketchLocation", ctx.SketchLocation)
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
	opts.Set("fqbn", ctx.FQBN.String())
	opts.Set("runtime.ide.version", ctx.ArduinoAPIVersion)
	opts.Set("customBuildProperties", strings.Join(ctx.CustomBuildProperties, ","))
	opts.Set("additionalFiles", strings.Join(additionalFilesRelative, ","))
	return opts
}

// InjectBuildOptions fill the context with fields from a properties map
func (ctx *Context) InjectBuildOptions(opts *properties.Map) {
	ctx.HardwareDirs = paths.NewPathList(strings.Split(opts.Get("hardwareFolders"), ",")...)
	ctx.ToolsDirs = paths.NewPathList(strings.Split(opts.Get("toolsFolders"), ",")...)
	ctx.BuiltInLibrariesDirs = paths.NewPathList(strings.Split(opts.Get("builtInLibrariesFolders"), ",")...)
	ctx.OtherLibrariesDirs = paths.NewPathList(strings.Split(opts.Get("otherLibrariesFolders"), ",")...)
	ctx.SketchLocation = opts.GetPath("sketchLocation")
	fqbn, err := cores.ParseFQBN(opts.Get("fqbn"))
	if err != nil {
		i18n.ErrorfWithLogger(ctx.GetLogger(), "Error in FQBN: %s", err)
	}
	ctx.FQBN = fqbn
	ctx.ArduinoAPIVersion = opts.Get("runtime.ide.version")
	ctx.CustomBuildProperties = strings.Split(opts.Get("customBuildProperties"), ",")
}

// GetLogger returns a logger
func (ctx *Context) GetLogger() i18n.Logger {
	if ctx.logger == nil {
		return &i18n.HumanLogger{}
	}
	return ctx.logger
}

// SetLogger sets a logger
func (ctx *Context) SetLogger(l i18n.Logger) {
	ctx.logger = l
}
