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

package types

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesresolver"
	"github.com/arduino/arduino-cli/arduino/sketch"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
)

type ProgressStruct struct {
	Progress   float32
	StepAmount float32
	Parent     *ProgressStruct
}

func (p *ProgressStruct) AddSubSteps(steps int) {
	p.Parent = &ProgressStruct{
		Progress:   p.Progress,
		StepAmount: p.StepAmount,
		Parent:     p.Parent,
	}
	if p.StepAmount == 0.0 {
		p.StepAmount = 100.0
	}
	p.StepAmount /= float32(steps)
}

func (p *ProgressStruct) RemoveSubSteps() {
	p.Progress = p.Parent.Progress
	p.StepAmount = p.Parent.StepAmount
	p.Parent = p.Parent.Parent
}

func (p *ProgressStruct) CompleteStep() {
	p.Progress += p.StepAmount
}

// Context structure
type Context struct {
	// Build options
	HardwareDirs         paths.PathList
	BuiltInToolsDirs     paths.PathList
	BuiltInLibrariesDirs *paths.Path
	OtherLibrariesDirs   paths.PathList
	LibraryDirs          paths.PathList // List of paths pointing to individual library root folders
	WatchedLocations     paths.PathList
	FQBN                 *cores.FQBN
	Clean                bool

	// Build options are serialized here
	BuildOptionsJson         string
	BuildOptionsJsonPrevious string

	PackageManager *packagemanager.Explorer
	RequiredTools  []*cores.ToolRelease
	TargetBoard    *cores.Board
	TargetPackage  *cores.Package
	TargetPlatform *cores.PlatformRelease
	ActualPlatform *cores.PlatformRelease

	BuildProperties              *properties.Map
	BuildPath                    *paths.Path
	SketchBuildPath              *paths.Path
	CoreBuildPath                *paths.Path
	CoreBuildCachePath           *paths.Path
	CoreArchiveFilePath          *paths.Path
	CoreObjectsFiles             paths.PathList
	LibrariesBuildPath           *paths.Path
	LibrariesObjectFiles         paths.PathList
	PreprocPath                  *paths.Path
	SketchObjectFiles            paths.PathList
	IgnoreSketchFolderNameErrors bool

	Sketch        *sketch.Sketch
	WarningsLevel string

	// Arduino sketch (.ino) to C++ (.cpp) conversion steps:
	// 1. Concatenate *.ino files into a single merged source file -> SketchSourceMerged
	SketchSourceMerged string

	// Libraries handling
	LibrariesManager             *librariesmanager.LibrariesManager
	LibrariesResolver            *librariesresolver.Cpp
	ImportedLibraries            libraries.List
	LibrariesResolutionResults   map[string]LibraryResolutionResult
	IncludeFolders               paths.PathList
	UseCachedLibrariesResolution bool

	// C++ Parsing
	LineOffset int

	// Verbosity settings
	Verbose bool

	// Dry run, only create progress map
	Progress ProgressStruct
	// Send progress events to this callback
	ProgressCB rpc.TaskProgressCB

	// Custom build properties defined by user (line by line as "key=value" pairs)
	CustomBuildProperties []string

	// Reuse old tools since the backing storage didn't change
	CanUseCachedTools bool

	// Experimental: use arduino-preprocessor to create prototypes
	UseArduinoPreprocessor bool

	// Parallel processes
	Jobs int

	// Out and Err stream to redirect all output
	Stdout  io.Writer
	Stderr  io.Writer
	stdLock sync.Mutex

	// Sizer results
	ExecutableSectionsSize ExecutablesFileSections

	// Compilation Database to build/update
	CompilationDatabase *builder.CompilationDatabase
	// Set to true to skip build and produce only Compilation Database
	OnlyUpdateCompilationDatabase bool

	// Source code overrides (filename -> content map).
	// The provided source data is used instead of reading it from disk.
	// The keys of the map are paths relative to sketch folder.
	SourceOverride map[string]string
}

// ExecutableSectionSize represents a section of the executable output file
type ExecutableSectionSize struct {
	Name    string `json:"name"`
	Size    int    `json:"size"`
	MaxSize int    `json:"max_size"`
}

// ExecutablesFileSections is an array of ExecutablesFileSection
type ExecutablesFileSections []ExecutableSectionSize

// ToRPCExecutableSectionSizeArray transforms this array into a []*rpc.ExecutableSectionSize
func (s ExecutablesFileSections) ToRPCExecutableSectionSizeArray() []*rpc.ExecutableSectionSize {
	res := []*rpc.ExecutableSectionSize{}
	for _, section := range s {
		res = append(res, &rpc.ExecutableSectionSize{
			Name:    section.Name,
			Size:    int64(section.Size),
			MaxSize: int64(section.MaxSize),
		})
	}
	return res
}

func (ctx *Context) ExtractBuildOptions() *properties.Map {
	opts := properties.NewMap()
	opts.Set("hardwareFolders", strings.Join(ctx.HardwareDirs.AsStrings(), ","))
	opts.Set("builtInToolsFolders", strings.Join(ctx.BuiltInToolsDirs.AsStrings(), ","))
	if ctx.BuiltInLibrariesDirs != nil {
		opts.Set("builtInLibrariesFolders", ctx.BuiltInLibrariesDirs.String())
	}
	opts.Set("otherLibrariesFolders", strings.Join(ctx.OtherLibrariesDirs.AsStrings(), ","))
	opts.SetPath("sketchLocation", ctx.Sketch.FullPath)
	var additionalFilesRelative []string
	absPath := ctx.Sketch.FullPath.Parent()
	for _, f := range ctx.Sketch.AdditionalFiles {
		relPath, err := f.RelTo(absPath)
		if err != nil {
			continue // ignore
		}
		additionalFilesRelative = append(additionalFilesRelative, relPath.String())
	}
	opts.Set("fqbn", ctx.FQBN.String())
	opts.Set("customBuildProperties", strings.Join(ctx.CustomBuildProperties, ","))
	opts.Set("additionalFiles", strings.Join(additionalFilesRelative, ","))
	opts.Set("compiler.optimization_flags", ctx.BuildProperties.Get("compiler.optimization_flags"))
	return opts
}

func (ctx *Context) PushProgress() {
	if ctx.ProgressCB != nil {
		ctx.ProgressCB(&rpc.TaskProgress{Percent: ctx.Progress.Progress})
	}
}

func (ctx *Context) Info(msg string) {
	ctx.stdLock.Lock()
	if ctx.Stdout == nil {
		fmt.Fprintln(os.Stdout, msg)
	} else {
		fmt.Fprintln(ctx.Stdout, msg)
	}
	ctx.stdLock.Unlock()
}

func (ctx *Context) Warn(msg string) {
	ctx.stdLock.Lock()
	if ctx.Stderr == nil {
		fmt.Fprintln(os.Stderr, msg)
	} else {
		fmt.Fprintln(ctx.Stderr, msg)
	}
	ctx.stdLock.Unlock()
}

func (ctx *Context) WriteStdout(data []byte) (int, error) {
	ctx.stdLock.Lock()
	defer ctx.stdLock.Unlock()
	if ctx.Stdout == nil {
		return os.Stdout.Write(data)
	}
	return ctx.Stdout.Write(data)
}

func (ctx *Context) WriteStderr(data []byte) (int, error) {
	ctx.stdLock.Lock()
	defer ctx.stdLock.Unlock()
	if ctx.Stderr == nil {
		return os.Stderr.Write(data)
	}
	return ctx.Stderr.Write(data)
}
