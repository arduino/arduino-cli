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

package packagemanager

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packageindex"
	"github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-map"
	"github.com/sirupsen/logrus"
	"go.bug.st/relaxed-semver"
)

// PackageManager defines the superior oracle which understands all about
// Arduino Packages, how to parse them, download, and so on.
//
// The manager also keeps track of the status of the Packages (their Platform Releases, actually)
// installed in the system.
type PackageManager struct {
	Log      logrus.FieldLogger
	packages *cores.Packages

	IndexDir    *paths.Path
	PackagesDir *paths.Path
	DownloadDir *paths.Path
	TempDir     *paths.Path
}

// NewPackageManager returns a new instance of the PackageManager
func NewPackageManager(indexDir, packagesDir, downloadDir, tempDir *paths.Path) *PackageManager {
	return &PackageManager{
		Log:         logrus.StandardLogger(),
		packages:    cores.NewPackages(),
		IndexDir:    indexDir,
		PackagesDir: packagesDir,
		DownloadDir: downloadDir,
		TempDir:     tempDir,
	}
}

func (pm *PackageManager) Clear() {
	pm.packages = cores.NewPackages()
}

func (pm *PackageManager) GetPackages() *cores.Packages {
	return pm.packages
}

func (pm *PackageManager) FindPlatformReleaseProvidingBoardsWithVidPid(vid, pid string) []*cores.PlatformRelease {
	res := []*cores.PlatformRelease{}
	for _, targetPackage := range pm.packages.Packages {
		for _, targetPlatform := range targetPackage.Platforms {
			platformRelease := targetPlatform.GetLatestRelease()
			if platformRelease == nil {
				continue
			}
			for _, boardManifest := range platformRelease.BoardsManifest {
				if boardManifest.HasUsbID(vid, pid) {
					res = append(res, platformRelease)
					break
				}
			}
		}
	}
	return res
}

func (pm *PackageManager) FindBoardsWithVidPid(vid, pid string) []*cores.Board {
	res := []*cores.Board{}
	for _, targetPackage := range pm.packages.Packages {
		for _, targetPlatform := range targetPackage.Platforms {
			if platform := targetPlatform.GetInstalled(); platform != nil {
				for _, board := range platform.Boards {
					if board.HasUsbID(vid, pid) {
						res = append(res, board)
					}
				}
			}
		}
	}
	return res
}

func (pm *PackageManager) FindBoardsWithID(id string) []*cores.Board {
	res := []*cores.Board{}
	for _, targetPackage := range pm.packages.Packages {
		for _, targetPlatform := range targetPackage.Platforms {
			if platform := targetPlatform.GetInstalled(); platform != nil {
				for _, board := range platform.Boards {
					if board.BoardId == id {
						res = append(res, board)
					}
				}
			}
		}
	}
	return res
}

// FindBoardWithFQBN returns the board identified by the fqbn, or an error
func (pm *PackageManager) FindBoardWithFQBN(fqbnIn string) (*cores.Board, error) {
	fqbn, err := cores.ParseFQBN(fqbnIn)
	if err != nil {
		return nil, fmt.Errorf("parsing fqbn: %s", err)
	}

	_, _, board, _, _, err := pm.ResolveFQBN(fqbn)
	return board, err
}

// ResolveFQBN returns, in order:
// - the Package pointed by the fqbn
// - the PlatformRelease pointed by the fqbn
// - the Board pointed by the fqbn
// - the build properties for the board considering also the
//   configuration part of the fqbn
// - the PlatformRelease to be used for the build (if the board
//   requires a 3rd party core it may be different from the
//   PlatformRelease pointed by the fqbn)
// - an error if any of the above is not found
//
// In case of error the partial results found in the meantime are
// returned together with the error.
func (pm *PackageManager) ResolveFQBN(fqbn *cores.FQBN) (
	*cores.Package, *cores.PlatformRelease, *cores.Board,
	properties.Map, *cores.PlatformRelease, error) {

	// Find package
	targetPackage := pm.packages.Packages[fqbn.Package]
	if targetPackage == nil {
		return nil, nil, nil, nil, nil,
			errors.New("unknown package " + fqbn.Package)
	}

	// Find platform
	platform := targetPackage.Platforms[fqbn.PlatformArch]
	if platform == nil {
		return targetPackage, nil, nil, nil, nil,
			fmt.Errorf("unknown platform %s:%s", targetPackage, fqbn.PlatformArch)
	}
	platformRelease := platform.GetInstalled()
	if platformRelease == nil {
		return targetPackage, nil, nil, nil, nil,
			fmt.Errorf("platform %s is not installed", platformRelease)
	}

	// Find board
	board := platformRelease.Boards[fqbn.BoardID]
	if board == nil {
		return targetPackage, platformRelease, nil, nil, nil,
			fmt.Errorf("board %s:%s not found", platformRelease, fqbn.BoardID)
	}

	buildProperties, err := board.GetBuildProperties(fqbn.Configs)
	if err != nil {
		return targetPackage, platformRelease, board, nil, nil,
			fmt.Errorf("getting build properties for board %s: %s", board, err)
	}

	// Determine the platform used for the build (in case the board refers
	// to a core contained in another platform)
	buildPlatformRelease := platformRelease
	coreParts := strings.Split(buildProperties["build.core"], ":")
	if len(coreParts) > 1 {
		referredPackage := coreParts[1]
		buildPackage := pm.packages.Packages[referredPackage]
		if buildPackage == nil {
			return targetPackage, platformRelease, board, buildProperties, nil,
				fmt.Errorf("missing package %s:%s required for build", referredPackage, platform)
		}
		buildPlatformRelease = buildPackage.Platforms[fqbn.PlatformArch].GetInstalled()
	}

	// No errors... phew!
	return targetPackage, platformRelease, board, buildProperties, buildPlatformRelease, nil
}

// LoadPackageIndex loads a package index by looking up the local cached file from the specified URL
func (pm *PackageManager) LoadPackageIndex(URL *url.URL) error {
	indexPath := pm.IndexDir.Join(path.Base(URL.Path))
	index, err := packageindex.LoadIndex(indexPath)
	if err != nil {
		return fmt.Errorf("loading json index file %s: %s", indexPath, err)
	}

	index.MergeIntoPackages(pm.packages)
	return nil
}

// Package looks for the Package with the given name, returning a structure
// able to perform further operations on that given resource
func (pm *PackageManager) Package(name string) *PackageActions {
	//TODO: perhaps these 2 structure should be merged? cores.Packages vs pkgmgr??
	var err error
	thePackage := pm.packages.Packages[name]
	if thePackage == nil {
		err = fmt.Errorf("package '%s' not found", name)
	}
	return &PackageActions{
		aPackage:     thePackage,
		forwardError: err,
	}
}

// Actions that can be done on a Package

// PackageActions defines what actions can be performed on the specific Package
// It serves as a status container for the fluent APIs
type PackageActions struct {
	aPackage     *cores.Package
	forwardError error
}

// Tool looks for the Tool with the given name, returning a structure
// able to perform further operations on that given resource
func (pa *PackageActions) Tool(name string) *ToolActions {
	var tool *cores.Tool
	err := pa.forwardError
	if err == nil {
		tool = pa.aPackage.Tools[name]

		if tool == nil {
			err = fmt.Errorf("tool '%s' not found in package '%s'", name, pa.aPackage.Name)
		}
	}
	return &ToolActions{
		tool:         tool,
		forwardError: err,
	}
}

// END -- Actions that can be done on a Package

// Actions that can be done on a Tool

// ToolActions defines what actions can be performed on the specific Tool
// It serves as a status container for the fluent APIs
type ToolActions struct {
	tool         *cores.Tool
	forwardError error
}

// Get returns the final representation of the Tool
func (ta *ToolActions) Get() (*cores.Tool, error) {
	err := ta.forwardError
	if err == nil {
		return ta.tool, nil
	}
	return nil, err
}

// IsInstalled checks whether any release of the Tool is installed in the system
func (ta *ToolActions) IsInstalled() (bool, error) {
	if ta.forwardError != nil {
		return false, ta.forwardError
	}

	for _, release := range ta.tool.Releases {
		if release.IsInstalled() {
			return true, nil
		}
	}
	return false, nil
}

func (ta *ToolActions) Release(version *semver.RelaxedVersion) *ToolReleaseActions {
	if ta.forwardError != nil {
		return &ToolReleaseActions{forwardError: ta.forwardError}
	}
	release := ta.tool.GetRelease(version)
	if release == nil {
		return &ToolReleaseActions{forwardError: fmt.Errorf("release %s not found for tool %s", version, ta.tool.String())}
	}
	return &ToolReleaseActions{release: release}
}

// END -- Actions that can be done on a Tool

// ToolReleaseActions defines what actions can be performed on the specific ToolRelease
// It serves as a status container for the fluent APIs
type ToolReleaseActions struct {
	release      *cores.ToolRelease
	forwardError error
}

func (tr *ToolReleaseActions) Get() (*cores.ToolRelease, error) {
	if tr.forwardError != nil {
		return nil, tr.forwardError
	}
	return tr.release, nil
}

func (pm *PackageManager) GetAllInstalledToolsReleases() []*cores.ToolRelease {
	tools := []*cores.ToolRelease{}
	for _, targetPackage := range pm.packages.Packages {
		for _, tool := range targetPackage.Tools {
			for _, release := range tool.Releases {
				if release.IsInstalled() {
					tools = append(tools, release)
				}
			}
		}
	}
	return tools
}

func (pm *PackageManager) FindToolsRequiredForBoard(board *cores.Board) ([]*cores.ToolRelease, error) {
	// core := board.Properties["build.core"]

	platform := board.PlatformRelease

	// maps "PACKAGER:TOOL" => ToolRelease
	foundTools := map[string]*cores.ToolRelease{}

	// a Platform may not specify required tools (because it's a platform that comes from a
	// sketchbook/hardware dir without a package_index.json) then add all available tools
	for _, targetPackage := range pm.packages.Packages {
		for _, tool := range targetPackage.Tools {
			rel := tool.GetLatestInstalled()
			if rel != nil {
				foundTools[rel.Tool.String()] = rel
			}
		}
	}

	// replace the default tools above with the specific required by the current platform
	for _, toolDep := range platform.Dependencies {
		tool := pm.FindToolDependency(toolDep)
		if tool == nil {
			return nil, fmt.Errorf("tool release not found: %s", toolDep)
		}
		foundTools[tool.Tool.String()] = tool
	}

	requiredTools := []*cores.ToolRelease{}
	for _, toolRel := range foundTools {
		requiredTools = append(requiredTools, toolRel)
	}
	return requiredTools, nil
}

func (pm *PackageManager) FindToolDependency(dep *cores.ToolDependency) *cores.ToolRelease {
	toolRelease, err := pm.Package(dep.ToolPackager).Tool(dep.ToolName).Release(dep.ToolVersion).Get()
	if err != nil {
		return nil
	}
	return toolRelease
}
