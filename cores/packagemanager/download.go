package packagemanager

import (
	"os"
	"path/filepath"
	"fmt"

	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
	"github.com/bcmi-labs/arduino-cli/cores"
	"github.com/bcmi-labs/arduino-cli/common/releases"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/sirupsen/logrus"
	"github.com/bcmi-labs/arduino-cli/configs"
)

// PlatformReference represents a tuple to identify a Platform
type PlatformReference struct {
	Package              string // The package where this Platform belongs to.
	PlatformArchitecture string
	PlatformVersion      string
}

// FIXME: Make more generic and decouple the error print logic (that list should not exists;
// rather a failure @ the first package)

// findItemsToDownload takes a set of PlatformReference and returns a set of items to download and
// a set of outputs for non existing platforms.
func (pm *packageManager) FindItemsToDownload(items []PlatformReference) (
	[]*cores.PlatformRelease, []*cores.ToolRelease, []output.ProcessResult) {

	itemC := len(items)
	retPlatforms := []*cores.PlatformRelease{}
	retTools := []*cores.ToolRelease{}
	fails := make([]output.ProcessResult, 0, itemC)

	// value is not used, this map is only to check if an item is inside (set implementation)
	// see https://stackoverflow.com/questions/34018908/golang-why-dont-we-have-a-set-datastructure
	presenceMap := make(map[string]bool, itemC)

	for _, item := range items {
		pkg, exists := pm.packages.Packages[item.Package]
		if !exists {
			fails = append(fails, output.ProcessResult{
				ItemName: item.PlatformArchitecture,
				Error:    fmt.Sprintf("Package %s not found", item.Package),
			})
			continue
		}
		platform, exists := pkg.Platforms[item.PlatformArchitecture]
		if !exists {
			fails = append(fails, output.ProcessResult{
				ItemName: item.PlatformArchitecture,
				Error:    "Platform not found",
			})
			continue
		}

		_, exists = presenceMap[item.PlatformArchitecture]
		if exists { //skip
			continue
		}

		release := platform.GetVersion(item.PlatformVersion)
		if release == nil {
			fails = append(fails, output.ProcessResult{
				ItemName: item.PlatformArchitecture,
				Error:    fmt.Sprintf("Version %s Not Found", item.PlatformVersion),
			})
			continue
		}

		// replaces "latest" with latest version too
		toolDeps, err := pm.packages.GetDepsOfPlatformRelease(release)
		if err != nil {
			fails = append(fails, output.ProcessResult{
				ItemName: item.PlatformArchitecture,
				Error: fmt.Sprintf("Cannot get tool dependencies of plafotmr %s: %s",
					platform.Name, err.Error()),
			})
			continue
		}

		retPlatforms = append(retPlatforms, release)

		presenceMap[platform.Name] = true
		for _, tool := range toolDeps {
			if presenceMap[tool.Tool.Name] {
				continue
			}
			presenceMap[tool.Tool.Name] = true
			retTools = append(retTools, tool)
		}
	}
	return retPlatforms, retTools, fails
}

// FIXME: Refactor this download logic to uncouple it from the presentation layer
// All this stuff is pkgmgr responsibility for sure

func (pm *packageManager) DownloadToolReleaseArchives(tools []*cores.ToolRelease,
	results *output.CoreProcessResults) {

	downloads := map[string]*releases.DownloadResource{}
	for _, tool := range tools {
		resource := tool.GetCompatibleFlavour()
		if resource == nil {
			formatter.PrintError(fmt.Errorf("missing tool %s", tool), "A release of the tool is not available for your OS")
		}
		downloads[tool.Tool.Name+"@"+tool.Version] = tool.GetCompatibleFlavour()
	}
	logrus.Info("Downloading tools")
	results.Tools = append(results.Tools, pm.downloadStuff(downloads)...)
}

func (pm *packageManager) DownloadPlatformReleaseArchives(platforms []*cores.PlatformRelease,
	results *output.CoreProcessResults) {

	downloads := map[string]*releases.DownloadResource{}
	for _, platform := range platforms {
		downloads[platform.Platform.Package.Name+":"+platform.Platform.Name+"@"+platform.Version] = platform.Resource
	}

	logrus.Info("Downloading cores")
	results.Cores = append(results.Cores, pm.downloadStuff(downloads)...)
}

func (pm *packageManager) downloadStuff(downloads map[string]*releases.DownloadResource) []output.ProcessResult {

	var downloadProgressHandler releases.ParallelDownloadProgressHandler
	if pm.eventHandler != nil {
		downloadProgressHandler = pm.eventHandler.OnDownloadingSomething()
	}

	downloadRes := releases.ParallelDownload(downloads, false,
		downloadProgressHandler)
	return formatter.ExtractProcessResultsFromDownloadResults(downloads, downloadRes, "Downloaded")
}

func (pm *packageManager) InstallToolReleases(toolReleasesToDownload []*cores.ToolRelease,
	result *output.CoreProcessResults) error {

	for _, item := range toolReleasesToDownload {
		logrus.WithField("Package", item.Tool.Package.Name).
			WithField("Name", item.Tool.Name).
			WithField("Version", item.Version).
			Info("Installing tool")

		toolRoot, err := configs.ToolsFolder(item.Tool.Package.Name).Get()
		if err != nil {
			formatter.PrintError(err, "Cannot get tool install path, try again.")
			return err
		}
		possiblePath := filepath.Join(toolRoot, item.Tool.Name, item.Version)

		err = cores.InstallTool(possiblePath, item.GetCompatibleFlavour())
		var processResult output.ProcessResult
		if err != nil {
			if os.IsExist(err) {
				logrus.WithError(err).Warnf("Cannot install tool `%s`, it is already installed", item.Tool.Name)
				processResult = output.ProcessResult{
					ItemName: item.Tool.Name,
					Status:   "Already Installed",
					Path:     possiblePath,
				}
			} else {
				logrus.WithError(err).Warnf("Cannot install tool `%s`", item.Tool.Name)
				processResult = output.ProcessResult{
					ItemName: item.Tool.Name,
					Status:   "",
					Error:    err.Error(),
				}
			}
		} else {
			logrus.Info("Adding installed tool to final result")
			processResult = output.ProcessResult{
				ItemName: item.Tool.Name,
				Status:   "Installed",
				Path:     possiblePath,
			}
		}
		result.Cores = append(result.Cores, processResult)
	}
	return nil
}

func (pm *packageManager) InstallPlatformReleases(platformReleasesToDownload []*cores.PlatformRelease,
	outputResults *output.CoreProcessResults) error {

	for _, item := range platformReleasesToDownload {
		logrus.WithField("Package", item.Platform.Package.Name).
			WithField("Name", item.Platform.Name).
			WithField("Version", item.Version).
			Info("Installing core")

		coreRoot, err := configs.CoresFolder(item.Platform.Package.Name).Get()
		if err != nil {
			return err
		}
		possiblePath := filepath.Join(coreRoot, item.Platform.Architecture, item.Version)

		err = cores.InstallPlatform(possiblePath, item.Resource)
		var result output.ProcessResult
		if err != nil {
			if os.IsExist(err) {
				logrus.WithError(err).Warnf("Cannot install core `%s`, it is already installed", item.Platform.Name)
				result = output.ProcessResult{
					ItemName: item.Platform.Name,
					Status:   "Already Installed",
					Path:     possiblePath,
				}
			} else {
				logrus.WithError(err).Warnf("Cannot install core `%s`", item.Platform.Name)
				result = output.ProcessResult{
					ItemName: item.Platform.Name,
					Status:   "",
					Error:    err.Error(),
				}
			}
		} else {
			logrus.Info("Adding installed core to final result")

			result = output.ProcessResult{
				ItemName: item.Platform.Name,
				Status:   "Installed",
				Path:     possiblePath,
			}
		}
		outputResults.Cores = append(outputResults.Cores, result)
	}
	return nil
}
