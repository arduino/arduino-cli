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

package packagemanager

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packageindex"
	paths "github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/sirupsen/logrus"
	semver "go.bug.st/relaxed-semver"
)

// PackageManager defines the superior oracle which understands all about
// Arduino Packages, how to parse them, download, and so on.
//
// The manager also keeps track of the status of the Packages (their Platform Releases, actually)
// installed in the system.
type PackageManager struct {
	Log                    logrus.FieldLogger
	Packages               cores.Packages
	IndexDir               *paths.Path
	PackagesDir            *paths.Path
	DownloadDir            *paths.Path
	TempDir                *paths.Path
	CustomGlobalProperties *properties.Map
}

// NewPackageManager returns a new instance of the PackageManager
func NewPackageManager(indexDir, packagesDir, downloadDir, tempDir *paths.Path) *PackageManager {
	return &PackageManager{
		Log:                    logrus.StandardLogger(),
		Packages:               cores.NewPackages(),
		IndexDir:               indexDir,
		PackagesDir:            packagesDir,
		DownloadDir:            downloadDir,
		TempDir:                tempDir,
		CustomGlobalProperties: properties.NewMap(),
	}
}

// Clear resets the PackageManager to its initial state
func (pm *PackageManager) Clear() {
	pm.Packages = cores.NewPackages()
	pm.CustomGlobalProperties = properties.NewMap()
}

// FindPlatformReleaseProvidingBoardsWithVidPid FIXMEDOC
func (pm *PackageManager) FindPlatformReleaseProvidingBoardsWithVidPid(vid, pid string) []*cores.PlatformRelease {
	res := []*cores.PlatformRelease{}
	for _, targetPackage := range pm.Packages {
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

// FindBoardsWithVidPid FIXMEDOC
func (pm *PackageManager) FindBoardsWithVidPid(vid, pid string) []*cores.Board {
	res := []*cores.Board{}
	for _, targetPackage := range pm.Packages {
		for _, targetPlatform := range targetPackage.Platforms {
			if platform := pm.GetInstalledPlatformRelease(targetPlatform); platform != nil {
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

// FindBoardsWithID FIXMEDOC
func (pm *PackageManager) FindBoardsWithID(id string) []*cores.Board {
	res := []*cores.Board{}
	for _, targetPackage := range pm.Packages {
		for _, targetPlatform := range targetPackage.Platforms {
			if platform := pm.GetInstalledPlatformRelease(targetPlatform); platform != nil {
				for _, board := range platform.Boards {
					if board.BoardID == id {
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
//
// - the Package pointed by the fqbn
//
// - the PlatformRelease pointed by the fqbn
//
// - the Board pointed by the fqbn
//
// - the build properties for the board considering also the
// configuration part of the fqbn
//
// - the PlatformRelease to be used for the build (if the board
// requires a 3rd party core it may be different from the
// PlatformRelease pointed by the fqbn)
//
// - an error if any of the above is not found
//
// In case of error the partial results found in the meantime are
// returned together with the error.
func (pm *PackageManager) ResolveFQBN(fqbn *cores.FQBN) (
	*cores.Package, *cores.PlatformRelease, *cores.Board,
	*properties.Map, *cores.PlatformRelease, error) {

	// Find package
	targetPackage := pm.Packages[fqbn.Package]
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
	platformRelease := pm.GetInstalledPlatformRelease(platform)
	if platformRelease == nil {
		return targetPackage, nil, nil, nil, nil,
			fmt.Errorf("platform %s is not installed", platform)
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
	coreParts := strings.Split(buildProperties.Get("build.core"), ":")
	if len(coreParts) > 1 {
		referredPackage := coreParts[0]
		buildPackage := pm.Packages[referredPackage]
		if buildPackage == nil {
			return targetPackage, platformRelease, board, buildProperties, nil,
				fmt.Errorf("missing package %s referenced by board %s", referredPackage, fqbn)
		}
		buildPlatform := buildPackage.Platforms[fqbn.PlatformArch]
		if buildPlatform == nil {
			return targetPackage, platformRelease, board, buildProperties, nil,
				fmt.Errorf("missing platform %s:%s referenced by board %s", referredPackage, fqbn.PlatformArch, fqbn)
		}
		buildPlatformRelease = pm.GetInstalledPlatformRelease(buildPlatform)
		if buildPlatformRelease == nil {
			return targetPackage, platformRelease, board, buildProperties, nil,
				fmt.Errorf("missing platform release %s:%s referenced by board %s", referredPackage, fqbn.PlatformArch, fqbn)
		}
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

	for _, p := range index.Packages {
		p.URL = URL.String()
	}

	index.MergeIntoPackages(pm.Packages)
	return nil
}

// LoadPackageIndexFromFile load a package index from the specified file
func (pm *PackageManager) LoadPackageIndexFromFile(indexPath *paths.Path) (*packageindex.Index, error) {
	index, err := packageindex.LoadIndex(indexPath)
	if err != nil {
		return nil, fmt.Errorf("loading json index file %s: %s", indexPath, err)
	}

	index.MergeIntoPackages(pm.Packages)
	return index, nil
}

// Package looks for the Package with the given name, returning a structure
// able to perform further operations on that given resource
func (pm *PackageManager) Package(name string) *PackageActions {
	//TODO: perhaps these 2 structure should be merged? cores.Packages vs pkgmgr??
	var err error
	thePackage := pm.Packages[name]
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

// Release FIXMEDOC
func (ta *ToolActions) Release(version *semver.RelaxedVersion) *ToolReleaseActions {
	if ta.forwardError != nil {
		return &ToolReleaseActions{forwardError: ta.forwardError}
	}
	release := ta.tool.FindReleaseWithRelaxedVersion(version)
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

// Get FIXMEDOC
func (tr *ToolReleaseActions) Get() (*cores.ToolRelease, error) {
	if tr.forwardError != nil {
		return nil, tr.forwardError
	}
	return tr.release, nil
}

// GetInstalledPlatformRelease returns the PlatformRelease installed (it is chosen)
func (pm *PackageManager) GetInstalledPlatformRelease(platform *cores.Platform) *cores.PlatformRelease {
	releases := platform.GetAllInstalled()
	if len(releases) == 0 {
		return nil
	}

	debug := func(msg string, pl *cores.PlatformRelease) {
		pm.Log.WithField("bundle", pl.IsIDEBundled).
			WithField("version", pl.Version).
			WithField("managed", pm.IsManagedPlatformRelease(pl)).
			Debugf("%s: %s", msg, pl)
	}

	best := releases[0]
	bestIsManaged := pm.IsManagedPlatformRelease(best)
	debug("current best", best)

	for _, candidate := range releases[1:] {
		candidateIsManaged := pm.IsManagedPlatformRelease(candidate)
		debug("candidate", candidate)
		// TODO: Disentangle this algorithm and make it more straightforward
		if bestIsManaged == candidateIsManaged {
			if best.IsIDEBundled == candidate.IsIDEBundled {
				if candidate.Version.GreaterThan(best.Version) {
					best = candidate
				}
			}
			if best.IsIDEBundled && !candidate.IsIDEBundled {
				best = candidate
			}
		}
		if !bestIsManaged && candidateIsManaged {
			best = candidate
			bestIsManaged = true
		}
		debug("current best", best)
	}
	return best
}

// GetAllInstalledToolsReleases FIXMEDOC
func (pm *PackageManager) GetAllInstalledToolsReleases() []*cores.ToolRelease {
	tools := []*cores.ToolRelease{}
	for _, targetPackage := range pm.Packages {
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

// InstalledPlatformReleases returns all installed PlatformReleases. This function is
// useful to range all PlatformReleases in for loops.
func (pm *PackageManager) InstalledPlatformReleases() []*cores.PlatformRelease {
	platforms := []*cores.PlatformRelease{}
	for _, targetPackage := range pm.Packages {
		for _, platform := range targetPackage.Platforms {
			for _, release := range platform.GetAllInstalled() {
				platforms = append(platforms, release)
			}
		}
	}
	return platforms
}

// InstalledBoards returns all installed Boards. This function is useful to range
// all Boards in for loops.
func (pm *PackageManager) InstalledBoards() []*cores.Board {
	boards := []*cores.Board{}
	for _, targetPackage := range pm.Packages {
		for _, platform := range targetPackage.Platforms {
			for _, release := range platform.GetAllInstalled() {
				for _, board := range release.Boards {
					boards = append(boards, board)
				}
			}
		}
	}
	return boards
}

// FindToolsRequiredForBoard FIXMEDOC
func (pm *PackageManager) FindToolsRequiredForBoard(board *cores.Board) ([]*cores.ToolRelease, error) {
	pm.Log.Infof("Searching tools required for board %s", board)

	// core := board.Properties["build.core"]
	platform := board.PlatformRelease

	// maps "PACKAGER:TOOL" => ToolRelease
	foundTools := map[string]*cores.ToolRelease{}

	// a Platform may not specify required tools (because it's a platform that comes from a
	// user/hardware dir without a package_index.json) then add all available tools
	for _, targetPackage := range pm.Packages {
		for _, tool := range targetPackage.Tools {
			rel := tool.GetLatestInstalled()
			if rel != nil {
				foundTools[rel.Tool.Name] = rel
			}
		}
	}

	// replace the default tools above with the specific required by the current platform
	requiredTools := []*cores.ToolRelease{}
	platform.Dependencies.Sort()
	for _, toolDep := range platform.Dependencies {
		pm.Log.WithField("tool", toolDep).Infof("Required tool")
		tool := pm.FindToolDependency(toolDep)
		if tool == nil {
			return nil, fmt.Errorf("tool release not found: %s", toolDep)
		}
		requiredTools = append(requiredTools, tool)
		delete(foundTools, tool.Tool.Name)
	}

	for _, toolRel := range foundTools {
		requiredTools = append(requiredTools, toolRel)
	}
	return requiredTools, nil
}

// FindToolDependency returns the ToolRelease referenced by the ToolDependency or nil if
// the referenced tool doesn't exists.
func (pm *PackageManager) FindToolDependency(dep *cores.ToolDependency) *cores.ToolRelease {
	toolRelease, err := pm.Package(dep.ToolPackager).Tool(dep.ToolName).Release(dep.ToolVersion).Get()
	if err != nil {
		return nil
	}
	return toolRelease
}
